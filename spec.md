# Staking Proposals

This document is intended to lay out the implementation order of various
staking features. Each iteration of the staking module should be deployable 
as a test-net. 

Overall the Staking module should be rolled out in the following steps

1. Self-Bonding
2. Delegation
4. Unbonding Period, Re-Delegation
5. Validation Provision Atoms
6. Fee Pool, Commission

## Self-Bonding

The majority of the processing will happen from the set of all Validator Bonds
which is defined as an array and persisted in the application state.  Each
validator bond is defined as the object below. 

``` golang
type ValidatorBond struct {
	PubKey       crypto.PubKey  // validator PubKey
	Owner        auth.Account   // account coins are bonded from and unbonded to
	BondedTokens uint64         // total number of bond tokens from the validator 
	VotingPower  uint64         // voting power for tendermint consensus 
	Description  Description    // arbitrary information for the UI 
}

type Description struct {
	Name     string `json:"pubkey"`
	Identity string `json:"identity"`
	Website  string `json:"website"`
	Details  string `json:"details"`
}
```

Note that the Validator PubKey is not necessarily the PubKey of the sender of
the transaction, but is defined by the transaction. This way a separated key
can be used for holding tokens and participating in consensus.

The VotingPower is proportional to the amount of bonded tokens which the validator
has if the validator is within the top 100 validators. At the launch of 
cosmos hub we will have a maximum of 100 validators. 

Within the store, each `ValidatorBond` is stored by validator-pubkey.

 - key: validator-pubkey
 - value: `ValidatorBond` object

A second key-value pair is also maintained by the store in order to quickly
sort though the the group of all validator-candidates: 

 - key: `ValidatorBond.BondedTokens`
 - value: `ValidatorBond.PubKey`

When the set of all validators needs to be determined from the group of all
candidates, the top 100 candidates, sorted by BondedTokens can be retrieved
from this sorting without the need to retrieve the entire group of candidates. 

BondedTokens are held in a global account.  

``` golang
type Param struct {
	BondedTokenPool    uint64     // reserve of all bonded tokens  
	HoldAccountBonded auth.Account  // Protocol account for bonded tokens 
}
```

TxBond is the self-bond SDK transaction for self-bonding tokens to a PubKey
which you designate.  Here the PubKey bytes are the serialized bytes of the
PubKey struct as defined by `tendermint/go-crypto`.

``` golang
type TxBond struct {
	PubKey       crypto.PubKey 
	Amount       coin.Coin
    Description  Description     
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
	Owner        auth.Account
	Shares       uint64    
	VotingPower  uint64   
    Description  Description     
}
```

At this phase in development the number of shares issued to each account will
be equal to the number of coins bonded to this candidate. In later phases when
staking provisions is introduced, the shares will not be equal to the number of
coins bonded to the candidate. 

DelegatorBond represents some bond shares held by an account. It is owned by
one delegator, and is associated with the voting power of one delegatee. The
sender of the transaction is considered to be the owner of the bond being sent. 

``` golang
type DelegatorBond struct {
	PubKey crypto.PubKey
	Shares uint64
} 
```

Within the store each `DelegatorBond` is individually stored by candidate
pubkey and sender actor.  

 - key: Delegator/Candidate-Pubkey
 - value: DelegatorBond 

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
    Description Description
}

type TxUnbond struct { 
	PubKey crypto.PubKey
	Shares uint64       
}
```

## Unbonding Period, Re-Delegation

### Unbonding

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
	Payout           auth.Account // account to pay out to
    Amount           uint64    // amount of shares which are unbonding
    StartSlashRatio  uint64    // old candidate slash ratio at start of re-delegation
}
```

The queue is ordered so the next to unbond is at the head.  Every tick the head
of the queue is checked and if the unbonding period has passed since
`InitHeight` commence with final settlment of the unbonding and pop the queue.
Currently the each share to unbond is equivelent to a single unbonding token,
although in later phases shares are used along with other terms to calculate
the coins.

When unbonding is initiated, delegator shares are immediately removed from the
candidate. In the unbonding queue - the fraction of all historical slashings on
that validator are recorded (StartSlashRatio). When this queue reaches maturity
if that total slashing applied is greater on the validator then the
difference (amount that should have been slashed from the first validator) is
assigned to the amount being paid out. 

### Re-Delegation

If a delegator is attempting to switch validators, using an unbonding command
followed by a bond command will reduce the amount of validation reward than
having just stayed with the same validator. Because we do not want to
disincentive delegators from switching validators, the re-delegate command is
introduced. This command provides a delegator who wishes to switch validators
equal reward to as if you had never unbonded. Transactions structs of the
re-delegation command may looks like this:

Re-delegation is defined as an unbond followed by an immediate bond at the 
maturity of the unbonding. 

```
type QueueElemReDelegate struct {
	QueueElem
	Payout       auth.Account  // account to pay out to
    Shares       uint64        // amount of shares which are unbonding
    NewCandidate crypto.PubKey // validator to bond to after unbond
}
```

When re-delegation is initiated, delegator shares remain accounted for within
the `Candidate.Shares` and the term `ReDelegatingShares` is incremented.
This must be done in anticipation for the next phase of implementation where we
must properly calculate payouts to unbonding token holders.  During the
unbonding period all unbonding shares do not count towards the voting power of
a validator. Once the `QueueElemReDelegation` has reached maturity, the
appropriate unbonding shares are removed from the `Shares` and
`ReDelegatingShares` term.   

As a unique case, any delegator who is in the process of unbonding from a
validator can use this transaction type to re-delegate back to the original
validator they're currently unbonding from (and only that validator).  This can
be thought of as a _cancel unbonding_ option.

### Candidate Unbonding

Additionally within this phase unbonding of entire candidates to temporary
liquid accounts is introduced. This full candidate unbond is used when a
candidate unbonds all of their tokens, is kicked from the validator set for
liveliness issues, or crosses a self-imposed safety threshold (defined in later
sections). Whenever a validator is kicked, the delegators do not automatically
unbond to their personal wallets, but unbond to a pool where they must transact
to completely withdraw their tokens. If a validator is able to regain
liveliness and participate in consensus once again, any delegator who has not
yet withdrawn from this liquid pool associated with that validator will
automatically be bonded back to the now-live validator. Mechanisms are also in
place to ensure that a delegator can continue to unbond, or re-delegate
instead, so long as the delegator acts in time.

``` golang
type QueueElemUnbondCandidate struct {
	QueueElem
}
```

A new parameter (`Status`) is now introduced into the Candidate struct to
signal that the candidate is either active, unbonding, or unbonded. Also
`UnbondingDelegatorShares` is introduced to account for delegators unbonding.  

``` golang
type CandidateStatus byte

const Active CandidateStatus = 0x00
const Unbonding CandidateStatus = 0x01
const Unbonded CandidateStatus = 0x02

type Candidate struct {
	Status             CandidateStatus
	PubKey             crypto.PubKey
	Owner              auth.Account
	Shares             uint64    
	ReDelegatingShares uint64  // Delegator shares currently re-delegating
	VotingPower        uint64   
    SlashRatio         uint64  // multiplicative slash ratio for this candidate
    Description        Description
}
```

At this point `TxDeclareCandidacy` can be used to reinstate an unbonding or
unbonded validator while providing any additionally required tokens. When this
occurs all previously delegated bonds will be applied to this validator
provided each delegator has not already withdrawn from the liquid pool, or
initiated another transaction from the validator

If a delegator chooses to initiate an unbond or re-delegation of their shares
while a candidate full unbond is commencing, then that unbond/re-delegation is
subject to a reduced unbonding period based on how much time that bond has
already spent in the unbonding queue.

## Validation Provision Atoms

In this phase validation provision tokens are introduced as the incentive
mechanism for tokens holders to keep their tokens bonded. All token holders are
incentivized to keep their tokens bonded as newly minted token are provided to
the bonded token supply. These token provisions are assigned proportionally to
the each slashable bonded token, this includes unbonding tokens. The intention
of validation previsions is not to increase the proportion of tokens held by
validators as opposed to delegators, therefor validators are not permitted to
charge validators a commission on their stake of the validation previsions.

Validation provisions are payed directly to a global hold account
(`BondedTokenPool`) and proportions of that hold account owned by each validator
is defined as the `GlobalStakeBonded`. The tokens are payed as bonded tokens.

``` golang
type Param struct {
	IssuedGlobalStakeShares  uint64  // sum of all the validators bonded shares
	BondedTokenPool           uint64  // reserve of all bonded tokens  
	HoldAccountBonded        auth.Account  // Protocol account for bonded tokens 
}
```

The candidate struct must now be expanded:

``` golang
type Candidate struct {
	Status                    byte       
	PubKey                    crypto.PubKey
	Owner                     auth.Account
	IssuedDelegatorShares     uint64    
	UnbondingDelegatorShares  uint64    
	GlobalStakeShares         uint64  
	VotingPower               uint64   
    Description               Description
}
```

The delegator struct will also need to account for the last liquid withdrawal:

``` golang
type DelegatorBond struct {
	PubKey crypto.PubKey
	Shares uint64
} 
```

Here, the bonded tokens that a candidate has can be calculated as:

```
globalStakeExRate = params.BondedTokenPool / params.IssuedGlobalStakeShares
candidateCoins = candidate.GlobalStakeShares * globalStakeExRate 
```

If a delegator chooses to add more tokens to a validator then the amount of
validator shares distributed is calculated on exchange rate (aka every
delegators shares do not change value at that moment. The validator's
accounting of distributed shares to delegators must also increased at every
deposit.
 
```
delegatorExRate = validatorCoins / candidate.IssuedDelegatorShares 
createShares = coinsDeposited / delegatorExRate 
candidate.IssuedDelegatorShares += createShares
```

Whenever a validator has new tokens added to it, the `BondedTokenPool` is
increased and must be reflected in the global parameter as well as the
validators `GlobalStakeShares`.  This calculation ensures that the worth of
the `GlobalStakeShares` of other validators remains worth a constant absolute
amount of the `BondedTokenPool`

```
createdGlobalStakeShares = coinsDeposited / globalStakeExRate 
validator.GlobalStakeShares +=  createdGlobalStakeShares
params.IssuedGlobalStakeShares +=  createdGlobalStakeShares

params.BondedTokenPool += coinsDeposited
```

Similarly, if a delegator wanted to unbond coins:

```
coinsWithdrawn = withdrawlShares * delegatorExRate

destroyedGlobalStakeShares = coinsWithdrawn / globalStakeExRate 
validator.GlobalStakeShares -= destroyedGlobalStakeShares
params.IssuedGlobalStakeShares -= destroyedGlobalStakeShares
params.BondedTokenPool -= coinsWithdrawn
```

Note that when an unbond occurs the shares to withdraw are placed in an
unbonding queue where they continue to collect validator provisions until queue
element matures. Although provisions are collected during unbonding unbonding
tokens do not contribute to the voting power of a validator. 

Validator provisions are minted on an hourly basis (the first block of a new
hour).  The annual target of between 7% and 20%. The long-term target ratio of
bonded tokens to unbonded tokens is 67%.  

The target annual inflation rate is recalculated for each previsions cycle. The
inflation is also subject to a rate change (positive of negative) depending or
the distance from the desired ratio (67%). The maximum rate change possible is
defined to be 13% per year, however the annual inflation is capped as between
7% and 20%.

```
inflationRateChange(0) = 0
annualInflation(0) = 0.07

bondedRatio = bondedTokenPool / totalTokenSupply
AnnualInflationRateChange = (1 - bondedRatio / 0.67) * 0.13

annualInflation += AnnualInflationRateChange

if annualInflation > 0.20 then annualInflation = 0.20
if annualInflation < 0.07 then annualInflation = 0.07

provisionTokensHourly = totalTokenSupply * annualInflation / (365.25*24)
```

Because the validators hold a relative bonded share (`GlobalStakeShare`), when
more bonded tokens are added proportionally to all validators the only term
which needs to be updated is the `BondedTokenPool`. So for each previsions cycle:

```
params.BondedTokenPool += provisionTokensHourly
```

## Fee Pool, Commission

Collected fees are pooled globally and divided out actively to validators and
passively to delegators. Each validator will now have the opportunity to charge
commission to the delegators on the fees collected on behalf of the deligators
by the validators. The presence of a commission mechanism creates
incentivization for the validators to run validator nodes as opposed to just
purely delegating to others. Fees are paid directly into a global fee pool 
 
``` golang
type Candidate struct {
	Status                 byte       
	PubKey                 crypto.PubKey
	Owner                  auth.Account
	IssuedDelegatorShares  uint64    
	GlobalStakeShares      uint64  
	VotingPower            uint64   
    Commission             uint64
    CommissionMax          uint64
    CommissionChangeRate   uint64
    CommissionChangeToday  int64
    FeeShares              uint64
    FeeCommissionShares    uint64
    LastFeesHeight         uint64
    LastFeesStakedShares   uint64
    Description            Description
}
```

Here several new parameters are introduced:
 - Commission:  The commission rate of fees charged to any delegators
 - CommissionMax:  The maximum commission rate which this candidate can charge
 - CommissionChangeRate: The maximum daily increase of the candidate commission
 - CommissionChangeToday: Counter for the amount of change to commission rate 
   which has occurred today, reset on the first block of each day (UTC time)
 - FeeShares: Cumulative counter for the amount of fees the candidate and its
   deligators are entitled too
 - FeeCommissionShares: Fee shares reserved for the candidate charged as
   commission to the delegators
 - LastFeesHeight: The last height fees moved from the `FeeHolding` to
   `FeePool` for this candidate
 - LastFeesStakedShares: The candidates staked shares at the time that the last
   fees were processed (at `LastFeesHeight`)

Similarly, the delegator bond must integrate an accum for fee pool accounting. 
Rather than individually account the accum, the accum may be calculated lazily 
based on the Height of the previous withdrawal.

``` golang
type DelegatorBond struct {
	Candidate           crypto.PubKey
	Shares              uint64
    FeeWithdrawalHeight uint64 // last height fees were withdrawn from Candidate.FeeShares
} 
```

`TxDeclareCandidacy` should be updated to include new relevant terms:

``` golang
type TxDeclareCandidacy struct {
	BondUpdate
    Description         Description
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
	IssuedGlobalStakeShares  uint64       // sum of all the validators bonded shares
	IssuedFeeShares          uint64       // sum of all fee pool shares issued 
	BondedTokenPool          uint64       // reserve of all bonded tokens  
	FeePool                  coin.Coins   // fee pool for all the fee shares which have already been distributed
	FeePoolShares            uint64       // total accumulated shares created for the fee pool
	FeeHoldings              coin.Coins   // collected fees waiting be added to the pool when assigned
	FeeHoldingsShares        uint64       // total accumulated shares created for the fee holdings
	FeeHoldingsBondedTokens  uint64       // total bonded tokens of validators considered when adding shares to FeeHoldings
	HoldAccountBonded        auth.Account // Protocol account for bonded tokens 
	HoldAccountFeePool       auth.Account // Protocol account for the fee pool 
	DateLastCommissionReset  uint64       // Unix Timestamp for last commission accounting reset
}
```

Some basic rules for the fee pool are as following:

 - When a either a delegator or a validator is withdrawing from the fee pool
   they must withdrawal the maximum amount they are entitled too
 - When bonding or unbonding tokens to an existing account  a full withdrawal of
   the fees must occur (as the rules for lazy accounting change)

Here a separate fee pool exists per candidate containing all fee asset held by
a validator. The candidate `FeeShares` are passively calculated for each block
as fees are collected.  When the validator is the proposer of a block then the
passive calculation of the fee pool shares takes place.  Additionally, when the
validator is the proposer of the round then a skew factor is provided to this
block give this validator an incentive to provide non-empty blocks. The skew
factor is calculated from precommits Tendermint messages and will be 1 for
non-proposer nodes and between 1.01 and 1.05 for proposer nodes.  

```
skew = 1.01 + 0.04 * number_precommits/number_validators
```

Fees are passively calculated using a second temporary pool dubbed as the
holding pool. For each proposer block, the fees are withdrawn for the proposer
validator, and the fee holding pool is increased to contain the fees for the
non-proposer validators to be withdrawn at a later block. 

```
proposerFeeCalcShares = skew * candidate.GlobalStakeShares
totalCalcShares = params.IssuedGlobalStakeShares + (skew - 1) 
                  * candidate.GlobalStakeShares 
ProposerBlockFees = FeesJustCollected * proposerFeeCalcShares / totalCalcShares

heights = (CurrentHeight - 1) - candidate.LastFeesHeight 
FeeHoldingSharesWithdrawn = heights * candidate.LastFeesStakedShares 
FeeHoldingWithdrawn = params.FeeHoldings 
                      * FeeHoldingSharesWithdrawn / params.FeeHoldingsShares

params.FeeHoldingsShares += totalCalcShares - proposerFeeCalcShares 
                            - FeeHoldingSharesWithdrawn
params.FeeHoldings -= FeesJustCollected - ProposerBlockFees 
                      - FeeHoldingWithdrawn
```
Add the fees to the pool and increase proposer pool shares. Note that shares
are divided into shares for delegators and shares collected as commission by
the validator
```
issuedShares = FeesForProposer * candidate.proposerFeeCalcShares
               + FeeHoldingWithdraw * candidate.LastFeesStakedShares

params.FeePool += ProposerBlockFees + FeeHoldingWithdrawn
params.FeePoolShares += issuedShares 

candidate.FeeShares += issuedShares * (1 - candidate.Commission)
candidate.FeeCommissionShares += issuedShares * candidate.Commission
params.IssuedFeeShares += issuedShares
```
lastly, reset the candidates `LastFeesStakedShares` and `LastFeesHeight`
```
candidate.LastFeesHeight = currentHeight
candidate.LastFeesStakedShares = candidate.GlobalStakeShares
```

To a delegator to withdraw fees a transaction must be sent to the candidate.
The upon withdrawal the liquid tokens are sent to the delegators liquid
account.  If the delegator account is the candidate.Owner (aka the candidate
owner is with withdrawing fees) then both the entitled fees as well as the
commission fees will be withdrawn.

``` golang
type TxWithdrawFees struct {
	PubKey crypto.PubKey
}
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

coinsWithdrawn = param.FeePool * withdrawalEntitlement
param.FeePool -= coinsWithdrawn
delegator.Owner.Balance += coinsWithdrawn
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


coinsWithdrawn = param.FeePool * withdrawalEntitlement
param.FeePool -= coinsWithdrawn
candidate.Owner.Balance. += coinsWithdrawn
```

Finally, note that when a delegator chooses to unbond shares, fees continue to
accumulate until the unbond queue reaches maturity. At the block which the
queue reaches maturity and shares are unbonded all available fees are
simultaneously withdrawn. 
