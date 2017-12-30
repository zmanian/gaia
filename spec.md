# Staking Module

The staking module is tasked with various core staking functionality.
Ultimately it is responsible for: 
 - Declaration of candidacy for becoming a validator
 - Updating tendermint validating power to reflect slashible stake
 - Delegation and unbonding transactions 
 - Implementing unbonding period
 - Provisioning Atoms
 - Managing and distributing transaction fees
 - Providing the framework for validator commission on delegators

## Global State

`Params` and `GlobalState` represent the global persistent state of the gaia.
`Params` is intended to remain static whereas `GlobalState` is anticipated to
change each block. 

``` golang
type Params struct {
	HoldBonded   auth.Account // account  where all bonded coins are held
	HoldUnbonded auth.Account // account where all delegated but unbonded coins are held

	InflationRateChange rational.Rational // maximum annual change in inflation rate
	InflationMax        rational.Rational // maximum inflation rate
	InflationMin        rational.Rational // minimum inflation rate
	GoalBonded          rational.Rational // Goal of percent bonded atoms

	MaxVals          uint16  // maximum number of validators
	AllowedBondDenom string  // bondable coin denomination

	// gas costs for txs
	GasDeclareCandidacy int64 
	GasEditCandidacy    int64 
	GasDelegate         int64 
	GasRedelegate       int64 
	GasUnbond           int64 
}

type GlobalState struct {
	TotalSupply       int64              // total supply of atom tokens
	BondedShares      PoolShares         // sum of all shares distributed for the Bonded Pool
	UnbondedShares    PoolShares         // sum of all shares distributed for the Unbonded Pool
	BondedPool        int64              // reserve of bonded tokens
	UnbondedPool      int64              // reserve of unbonded tokens held with candidates
	InflationLastTime int64              // timestamp of last processing of inflation
	Inflation         rational.Rational  // current annual inflation rate
    DateLastCommissionReset  int64       // Unix Timestamp for last commission accounting reset
	
    FeePool                  coin.Coins  // fee pool for all the fee shares which have already been distributed
	FeePoolShares            int64       // total accumulated shares created for the fee pool
	FeeHoldings              coin.Coins  // collected fees waiting be added to the pool when assigned
	FeeHoldingsShares        int64       // total accumulated shares created for the fee holdings
	FeeHoldingsBondedTokens  int64       // total bonded tokens of validators considered when adding shares to FeeHoldings
}
```


## Candidacy

XXX insert Candidate introduction

type TxDeclareCandidacy struct {
	BondUpdate
    Description         Description
	Commission          int64  
	CommissionMax       int64 
	CommissionMaxChange int64 
}
``` golang
type CandidateStatus byte

const Active CandidateStatus = 0x00
const Unbonding CandidateStatus = 0x01
const Unbonded CandidateStatus = 0x02

type Candidate struct {
	Status                 CandidateStatus       
	PubKey                 crypto.PubKey
	Owner                  auth.Account
	IssuedDelegatorShares  int64    
	GlobalStakeShares      int64  
	VotingPower            int64   
    Commission             int64
    CommissionMax          int64
    CommissionChangeRate   int64
    CommissionChangeToday  int64
    FeeShares              int64
    FeeCommissionShares    int64
    LastFeesHeight         int64
    LastFeesStakedShares   int64
    Description            Description
}
	PubKey       crypto.PubKey  // validator PubKey
	Owner        auth.Account   // account coins are bonded from and unbonded to
	BondedTokens int64         // total number of bond tokens from the validator 
	VotingPower  int64         // voting power for tendermint consensus 
	Description  Description    // arbitrary information for the UI 

type Description struct {
	Name       string 
	DateBonded string 
	Identity   string 
	Website    string 
	Details    string 
}
```

Candidate parameters are described:
 - Status: signal that the candidate is either active, unbonding, or unbonded.
   Also `UnbondingDelegatorShares` is account for delegators unbonding
 - PubKey: separated key from the owner of the candidate as is used strictly
   for participating in consensus.
 - Commission:  The commission rate of fees charged to any delegators
 - CommissionMax:  The maximum commission rate which this candidate can charge
 - CommissionChangeRate: The maximum daily increase of the candidate commission
 - CommissionChangeToday: Counter for the amount of change to commission rate 
   which has occurred today, reset on the first block of each day (UTC time)
 - FeeShares: Cumulative counter for the amount of fees the candidate and its
   delegators are entitled too
 - FeeCommissionShares: Fee shares reserved for the candidate charged as
   commission to the delegators
 - LastFeesHeight: The last height fees moved from the `FeeHolding` to
   `FeePool` for this candidate
 - LastFeesStakedShares: The candidates staked shares at the time that the last
   fees were processed (at `LastFeesHeight`) 
 - VotingPower: Proportional to the amount of bonded tokens which the validator
   has if the validator is within the top 100 validators.


### Notes on store indexing

Within the store, each `Candidate` is stored by validator-pubkey.

 - key: validator-pubkey
 - value: `ValidatorBond` object

A second key-value pair is also maintained by the store in order to quickly
sort though the the group of all validator-candidates: 

 - key: `ValidatorBond.BondedTokens`
 - value: `ValidatorBond.PubKey`

When the set of all validators needs to be determined from the group of all
candidates, the top 100 candidates, sorted by BondedTokens can be retrieved
from this sorting without the need to retrieve the entire group of candidates. 



### The Queue

In order to process the unbonding period a queue is defined which creates

``` golang
type Queue interface {
    Push(bytes []byte) // add a new entry to the end of the queue
    Pop()              // remove the leading queue entry
    Peek() []byte      // get the leading queue entry
}

type MerkleQueue struct {
	slot  byte           // queue slot in the store
	store state.SimpleDB // queue store
	tail  int64          // start position of the queue
	head  int64          // end position of the queue
}
```

All queue elements used for unbonding share a common struct:

``` golang
type QueueElem struct {
	Candidate   crypto.PubKey
	InitHeight  int64    // when the queue was initiated
}
```

## Delegator Bonding

``` golang 
type BondUpdate struct {
	PubKey crypto.PubKey
	Amount coin.Coin       
}

type TxDelegate struct { 
    BondUpdate 
}
```

For all subsequent bonding, whether self-bonding or delegation the `TxDelegate`
function should be used. In this context `TxUnbond` is used to unbond either
delegation bonds or validator self-bonds. 

Similarly, the delegator bond must integrate an accum for fee pool accounting. 
Rather than individually account the accum, the accum may be calculated lazily 
based on the Height of the previous withdrawal.

``` golang
type DelegatorBond struct {
	Candidate           crypto.PubKey
	Shares              rational.Rat
    FeeWithdrawalHeight int64  
} 
```

Description: 
 - Candidate: bond candidate pub-key
 - Shares: the number of shares associated with the pubkey, The shares will not
   be equal to the number of coins bonded to the candidate. 
 - FeeWithdrawalHeight: last height fees were withdrawn from
   `Candidate.FeeShares`

Each `DelegatorBond` is individually stored by candidate pubkey and sender
actor.  

 - key: Delegator/Candidate-Pubkey
 - value: DelegatorBond 


DelegatorBond represents some bond shares held by an account. It is owned by
one delegator, and is associated with the voting power of one validator. The
sender of the transaction is considered to be the owner of the bond being sent. 

### Delegator Unbonding

Delegator unbonding is defined by the following trasaction type:

``` golang
type TxUnbond struct { 
	PubKey crypto.PubKey
	Shares rational.Rat 
}
```

When unbonding is initiated, delegator shares are immediately removed from the
candidate and added to a queue object.

``` golang
type QueueElemUnbondDelegation struct {
	QueueElem
	Payout           auth.Account  // account to pay out to
    Shares           rational.Rat  // amount of shares which are unbonding
    StartSlashRatio  rational.Rat  // candidate slash ratio at start of re-delegation
}
```

In the unbonding queue - the fraction of all historical slashings on
that validator are recorded (StartSlashRatio). When this queue reaches maturity
if that total slashing applied is greater on the validator then the
difference (amount that should have been slashed from the first validator) is
assigned to the amount being paid out. 

The queue is ordered so the next to unbond is at the head. Every tick the head
of the queue is checked and if the unbonding period has passed since
`InitHeight` commence with final settlment of the unbonding and pop the queue.

### Re-Delegation

``` golang
type QueueElemReDelegate struct {
	QueueElem
	Payout       auth.Account  // account to pay out to
    Shares       rational.Rat  // amount of shares which are unbonding
    NewCandidate crypto.PubKey // validator to bond to after unbond
}
```

The redelegtion command allows delegators to switch validators while still 
receiving equal reward to as if you had never unbonded. Transactions structs of the
re-delegation command may looks like this:

``` golang
type TxRedelegate struct { 
	PubKeyFrom crypto.PubKey
	PubKeyTo   crypto.PubKey
	Shares     rational.Rat 
}
```

Re-delegation is defined as an unbond followed by an immediate bond at the 
maturity of the unbonding. 

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

``` golang
type QueueElemUnbondCandidate struct {
	QueueElem
}
```

Unbonding of entire candidates to temporary liquid accounts occurs under the
scenarios: 
 - a candidate unbonds all of their tokens
 - is kicked from the validator set for liveliness issues
 - crosses a self-imposed safety threshold
   - XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
   - XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXxXxXX

Delegators do not unbond to their personal wallets but begin the unbonding
process to a pool where they must then transact in order to withdraw to their
respective wallets.  If a delegator chooses to initiate an unbond or
re-delegation of their shares while a candidate full unbond is commencing, then
that unbond/re-delegation is subject to a reduced unbonding period based on how
much time that bond has already spent in the unbonding queue.

If the validator was kicked for liviliness issues and is able to regain
livliness then all tokens in the unbonding pool which have not transacted to
move will be bonded back to the now-live validator. 



## Fee Calculations

``` golang
type TxWithdrawFees struct {
	PubKey crypto.PubKey
}
```

Collected fees are pooled globally and divided out actively to validators and
passively to delegators. Each validator has the opportunity to charge
commission to the delegators on the fees collected on behalf of the delegators
by the validators. Fees are paid directly into a global fee pool 

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

## Provision Calculations

Every hour atom provisions are assigned proportionally to the each slashable
bonded token, not including unbonding tokens.

Validation provisions are payed directly to a global hold account
(`BondedTokenPool`) and proportions of that hold account owned by each
validator is defined as the `GlobalStakeBonded`. The tokens are payed as bonded
tokens.

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

