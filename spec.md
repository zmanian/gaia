# Staking Proposals

This document is intended to lay out the implementation order of various
staking features. Each iteration of the staking module should be deployable 
as a test-net. Please contribute to this document and add discussion :)

Overall the Staking module should be rolled out in the following steps

1. Self-Bonding
2. Delegation
4. Unbonding period
5. Fee Pool, Commission
6. Delegator Rebonding

## Self-Bonding

The majority of the processing will happen from the set of all Validator Bonds
which is defined as an array and persisted in the application state.  Each
validator bond is defined as the object below. 

``` golang
type ValidatorBonds []*ValidatorBond

type ValidatorBond struct {
	PubKey       crypto.PubKey  // validator PubKey
	Owner        sdk.Actor      // account coins are bonded from and unbonded to
	BondedTokens uint64         // total number of bond tokens from the validator 
	VotingPower  uint64         // voting power for tendermint consensus 
	Notes        string         // arbitrary information for the UI 
}
```
Note that the Validator PubKey is not necessarily the PubKey of the sender of
the transaction, but is defined by the transaction. This way a separated key
can be used for holding tokens and participating in consensus. 

The VotingPower is proportional to the amount of bonded tokens which the validator
has if the validator is within the top 100 validators. At the launch of 
cosmos hub we will have a maximum of 100 validators.

TxBond is the self-bond SDK transaction for self-bonding tokens to a PubKey
which you designate.  Here the PubKey bytes are the serialized bytes of the
PubKey struct as defined by `tendermint/go-crypto`.

``` golang
type TxBond struct {
	PubKey crypto.PubKey 
	Amount coin.Coin
    Notes  string     
}
```

Similarity, `TxUnbond` can is used to unbond tokens from the validators bond
account. The key difference is from `TxBond` is that the PubKey does not need
to be included as it has already been associated with the sender account.


``` golang
type TxUnbond struct {
	Amount coin.Coin `json:"amount"`
}
```

## Delegation

The next phase of development includes delegation functionality. All
transactions are still instantaneous in this development level.

The previous struct `ValidatorBond` is now split up into two structs which
represent the candidate account (`Candidate`) or the bond to a candidate
account (`DelegatorBond`).

Each validator-candidate bond is defined as the object below. 

``` golang
type Candidate struct {
	PubKey       crypto.PubKey
	Owner        sdk.Actor
	Shares       uint64    
	VotingPower  uint64   
	Notes        string   
}
```

At this phase in development the number of shares issued to each account will
be equal to the number of coins bonded to this candidate. In later phases when
staking provisions is introduced, the shares will not be equal to the number of
coins bonded to the candidate. 

DelegatorBond represents some bond tokens held by an account. It is owned by
one delegator, and is associated with the voting power of one delegatee.

``` golang
type DelegatorBond struct {
	PubKey crypto.PubKey
	Shares uint64
} 
```

The sender of the transaction is considered to be the owner of the bond being
sent. 

The transactions types must be renamed. `TxBond` is now effectively
`TxDeclareCandidacy` and but is only used to become a new validator. For all
subsequent bonding, whether self-bonding or delegation the `TxDelegate`
function should be used. In this context `TxUnbond` is used to unbond either
delegation bonds or validator self-bonds. As such `TxUnbond` must now include
the PubKey which the unbonding is to occur from.  

``` golang
type BondUpdate struct {
	PubKey crypto.PubKey
	Amount coin.Coin       
}

type TxDelegate struct { 
    BondUpdate 
}

type TxDeclareCandidacy struct { 
    BondUpdate
    Notes string
}

type TxUnbond struct { 
    BondUpdate 
}
```

## Unbonding Period

A staking-module state parameter will be introduced which defines the number of
blocks to unbond from a validator. Once complete unbonding as a validator or a
Candidate will put your coins in a queue before being returned.

``` golang
type MerkleQueue struct {
	slot  byte           //Queue name in the store
	store state.SimpleDB //Queue store
	tail  uint64         //Start position of the queue
	head  uint64         //End position of the queue
}

type Queue interface {
    Push(bytes []byte) 
    Pop() 
    Peek() []byte 
}
```
The queue element struct for unbonding delegator bond should look like this: 

``` golang
type QueueElem struct {
	Candidate   crypto.PubKey
	InitHeight  uint64    // when the queue was initiated
}

type QueueElemUnbondDelegation struct {
	QueueElem
	Payout      sdk.Actor // account to pay out to
    Shares      uint64    // amount of shares which are unbonding
}
```

Additionally within this phase unbonding of entire candidates to temporary
liquid accounts is introduced. This full candidate unbond is used when a
candidate unbonds all of their atoms, is kicked from the validator set for
liveliness issues, or crosses a self-imposed safety threshold (defined in later
sections) 

``` golang
type QueueElemUnbondCandidate struct {
	QueueElem
}
```

An additional parameter (Status) must now be introduced into the Candidate
struct to deal with this unbonding and unbonded state of a candidate 

``` golang
type Candidate struct {
	Status       byte          // 0x00 active, 0x01 unbonding, 0x02 unbonded
	PubKey       crypto.PubKey
	Owner        sdk.Actor
	Shares       uint64    
	VotingPower  uint64   
	Notes        string   
}
```

While unbonding or unbonded a candidate nolong collects validation provisions.
Additionally a delegator may choose to initiate an unbond of their delegated
tokens while a candidate full unbond is commencing. This an unbond from an
unbonding candidate will already have completed some of the unbonding period
and is therefor subject to a reduced unbonding period when is is begun. 

Additionally at this point `TxDeclareCandidacy` can be used to reinstate an
unbonding or unbonded validator while providing any additionally required
atoms.  

## Validation Prevision Atoms

In this phase validation provision atoms are introduced as the incentive
mechanism for atoms holders to keep their atoms bonded.  All atom holders are
incentivized to keep their atoms bonded as newly minted atom are provided to
the bonded atom supply. These atom provisions are assigned  proportionally to
the each bonded atom.  The intention of validation previsions is not to increase
the proportion of atoms held by validators as opposed to delegators, therefor
validators are not permitted to charge validators a commission on their stake
of the validation previsions.

Validation provisions are payed directly to a global hold account
(`BondedAtomPool`) and proportions of that hold account owned by each validator
is defined as the `GlobalStakeBonded`. The atoms are payed as bonded atoms. The
The global params used to account for the fee pools are introduced in the
persistent object `Params`. 

``` golang
type Param struct {
	IssuedGlobalStakeShares  uint64  // sum of all the validators bonded shares
	BondedAtomPool           uint64  // reserve of all bonded atoms  
	HoldAccountBonded        sdk.Actor  // Protocol account for bonded atoms 
}
```

The candidate struct must now be expanded:

``` golang
type Candidate struct {
	Status                 byte       
	PubKey                 crypto.PubKey
	Owner                  sdk.Actor
	IssuedDelegatorShares  uint64    
	GlobalStakeShares      uint64  
	VotingPower            uint64   
	Notes                  string   
}
```

The delegator struct will also need to account for the last liquid withdrawal:

``` golang
type DelegatorBond struct {
	PubKey crypto.PubKey
	Shares uint64
} 
```

Here, the bonded atoms that a validator has can be calculated as:

```
globalStakeExRate = params.BondedAtomPool / params.IssuedGlobalStakeShares
validatorCoins = candidate.GlobalStakeShares * globalStakeExRate 
```

If a delegator chooses to add more coins to a validator then the amount of
validator shares distributed is calculated on exchange rate (aka every
delegators shares do not change value at that moment. The validator's
accounting of distributed shares to delegators must also increased at every
deposit.
 
```
delegatorExRate = validatorCoins / candidate.IssuedDelegatorShares 
createShares = coinsDeposited / delegatorExRate 
candidate.IssuedDelegatorShares += createShares
```

Whenever a validator has new tokens added to it, the `BondedAtomPool` is
increased and must be reflected in the global parameter as well as the
validators `GlobalStakeShares`.  This calculation ensures that the worth of
the `GlobalStakeShares` of other validators remains worth a constant absolute
amount of the `BondedAtomPool`

```
createdGlobalStakeShares += coinsDeposited / globalStakeExRate 
validator.GlobalStakeShares +=  createdGlobalStakeShares
params.IssuedGlobalStakeShares +=  createdGlobalStakeShares

params.BondedAtomPool += coinsDeposited
```

Similarly, if a delegator wanted to unbond coins:

```
coinsWithdrawn = withdrawlShares * delegatorExRate

destroyedGlobalStakeShares = coinsWithdrawn / globalStakeExRate 
validator.GlobalStakeShares -= destroyedGlobalStakeShares
params.IssuedGlobalStakeShares -= destroyedGlobalStakeShares
params.BondedAtomPool -= coinsWithdrawn
```

Validator provisions are minted on an hourly basis (the first block of a new
hour).  The annual target of between 7% and 20%. The long-term target ratio of
bonded atoms to unbonded atoms is 67%.  

The target annual inflation rate is recalculated for each previsions cycle. The
inflation is also subject to a rate change (positive of negative) depending or
the distance from the desired ratio (67%). The maximum rate change possible is
defined to be 13% per year, however the annual inflation is capped as between
7% and 20%.

```
inflationRateChange(0) = 0
annualInflation(0) = 0.07

bondedRatio = bondedAtomPool / totalAtomSupply
AnnualInflationRateChange = (1 - bondedRatio / 0.67) * 0.13

annualInflation += AnnualInflationRateChange

if annualInflation > 0.20 then annualInflation = 0.20
if annualInflation < 0.07 then annualInflation = 0.07

provisionAtomsHourly = totalAtomSupply * annualInflation / (365.25*24)
```

Because the validators hold a relative bonded share (`GlobalStakeShare`), when
more bonded atoms are added proportionally to all validators the only term
which needs to be updated is the `BondedAtomPool`. So for each previsions cycle:

```
params.BondedAtomPool += provisionAtomsHourly
```

## Fee Pool, Commission

Collected fees are pooled globally and paid out lazily to validators and
delegators. Each validator will now have the opportunity to charge commission
to the delegators on the fees collected on behalf of the deligators by the
validators. The presence of a commission mechanism creates incentivization for
the validators to run validator nodes as opposed to just purely delegating to
others. 

Validators must specify the rate and the maximum value of the commission
charged to delegators as an act of self-regulation.  
 
``` golang
type Candidate struct {
	Status                 byte       
	PubKey                 crypto.PubKey
	Owner                  sdk.Actor
	IssuedDelegatorShares  uint64    
	GlobalStakeShares      uint64  
	VotingPower            uint64   
    Commission             uint64
    CommissionMax          uint64
    CommissionChangeRate   uint64
    CommissionChangeToday  uint64
    FeeShares              uint64
    FeeCommissionShares    uint64
	Notes                  string   
}
```

Here several new parameters are introduced:
 - Commission:  The commission rate of fees charged to any delegators
 - CommissionMax:  The maximum commission rate which this candidate can charge
 - CommissionChangeRate: The maximum daily change of the candidate commission
 - CommissionChangeToday: Counter for the amount of change to commission rate 
   which has occured today, reset on the first block of each day (UTC time)
 - FeeShares: Cumulative counter for the amount of fees the candidate and its
   deligators are entitled too
 - FeeCommissionShares: Fee shares reserved for the candidate charged as
   commission to the delegators

Similarly, the delegator bond must integrate an accum for fee pool accounting. 
Rather than individually account the accum, the accum may be calculated lazily 
based on the Height of the previous withdrawal.

``` golang
type DelegatorBond struct {
	Candidate           crypto.PubKey
	Shares              uint64
    FeeWithdrawalHeight uint64 // last height fees were withdrawn from
} 
```

`TxDeclareCandidacy` should be updated to include new relevant terms:

``` golang
type TxDeclareCandidacy struct {
	BondUpdate
    Notes               string
	Commission          uint64  
	CommissionMax       uint64 
	CommissionMaxChange uint64 
}
```

The global parameters must be updated to include the fee pool, commission
accounting reset day,  the sum of all validator power which will be used for
calculation fee portions

``` golang
type Param struct {
	IssuedGlobalStakeShares  uint64  // sum of all the validators bonded shares
	IssuedFeeShares          uint64  // sum of all fee pool shares issued 
	BondedAtomPool           uint64  // reserve of all bonded atoms  
	FeePool                  coin.Coins // fee pool
	HoldAccountBonded        sdk.Actor  // Protocol account for bonded atoms 
	HoldAccountFeePool       sdk.Actor  // Protocol account for the fee pool 
	DateLastCommissionReset  uint64     // Unix Timestamp for last commission accounting reset
}
```

Some basic rules for the fee pool are as following:

 - When a either a delegator or a validator is withdrawing from the fee pool
   they must withdrawal the maximum amount they are entitled too
 - When bonding or unbonding atoms to an existing account  a full withdrawal of
   the fees must occur (as the rules for lazy accounting change)

Here a separate fee pool exists per candidate for every fee asset held by a
validator. The candidate accum increments each block as fees are collected.  If
the validator is the proposer of the round then a skew factor is provided to
give this validator an incentive to provide non-empty blocks. The skew factor
is passed in through Tendermint core and will be 1 for proposer nodes and
between 1.01 and 1.05 for proposer nodes.  

```
issuedShares = Skew * candidate.VotingPower / param.TotalVotingPower 
candidate.FeeShares += issuedShares * (1 - candidate.Commission/PRECISION)
candidate.FeeCommissionShares += issuedShares * candidate.Commission/PRECISION
params.IssuedFeeShares += issuedShares
```

The fee pool entitlement of each share (whether `FeeShares` or
`FeeCommissionShares`)  can be calculated as: 

```
entitlement = shares / params.IssuedFeeShares
```

The entitlement is applied equally to all assets held in the fee pool.
Similarly, the `withdrawalEntitlement` is the percent of all assets in the fee
pool to be withdrawn during a delegator fee withdrawal: 

```
feeSharesToWithdraw = (currentHeight - delegator.FeeWithdrawalHeight) 
                      * delegator.Shares / candidate.IssueDelegatorShares
withdrawalEntitlement := feeSharesToWithdraw / params.IssuedFeeShares
```

When a withdrawal of a delegator fees occurs the shares must also be balanced
appropriately:

```
candidate.FeeShares -= feeSharesToWithdraw 
params.IssuedFeeShares -= feeSharesToWithdraw
```

Similarly when a candidate chooses to withdraw fees from the FeeCommissionShares:

```
withdrawalEntitlement := candidate.FeeCommissionShares / params.IssuedFeeShares
candidate.FeeCommissionShares = 0  
params.IssuedFeeShares -= FeeCommissionShares
```

## Re-delegation

If a delegator is attempting to switch validators, using an unbonding command
followed by a bond command will reduce the amount of validation reward than
having just stayed with the same validator. Because we do not want to
disincentive delegators from switching validators, we may introduce a re-delegate
command which provides equal reward to as if you had never unbonded. 

The moment you are re-delegation the old validator loses your voting power, and
the new validator gains this voting power. For the duration of the unbonding
period any re-delegating account will be slashable by both the old validator and
the new validator. 

The mechanism for double slashability is as follows. When re-delegating the
hold account atoms are automatically moved from the old validator to the new
validator. Within the new validator - the atoms are treated as a regular
delegation. Within the old validator, the atoms have already been removed its
hold account but a separate mechanism applied to the re-delegation queue item -
in the re-delegating queue - the fraction of all historical slashings on that
validator are recorded. When this queue reaches maturity if that total
slashing applied is greater on the old validator then the difference (amount
that should have been slashed from the first validator) is assigned as debt to
the delegator addresses. If a delegator address has dept - the only operation
they are permitted to perform is an unbond operation, additionally they cannot
vote, or accrue and new fees from the time the debt was added.

Transactions structs of the re-delegation command may looks like this: 

``` golang 
type TxRedelegate struct { 
    BondUpdate 
    NewValidator crypto.PubKey 
} 
```

Within the delegation bond we will need to now add a debt as explained above:

``` golang
type DelegatorBond struct {
	PubKey           crypto.PubKey
	Shares           uint64
	Debt             uint64
    LastFeeWithdrawl time.Time
}
```

Lastly it should also be noted that the TxRedelegate function can be performed
during a candidate unbond as outlined in earlier phase _Unbonding Period_. This
would allow a delegator who is bonded to an inactive candidate to immediately
join a working validator. Any re-delegation from an unbonding candidate would
be subject to a smaller time in the redelegation queue. 
