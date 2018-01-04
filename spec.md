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
```

``` golang
type GlobalState struct {
	TotalSupply              int64        // total supply of atom tokens
	BondedShares             rational.Rat // sum of all shares distributed for the BondedPool
	UnbondedShares           rational.Rat // sum of all shares distributed for the UnbondedPool
	BondedPool               int64        // reserve of bonded tokens
	UnbondedPool             int64        // reserve of unbonded tokens held with candidates
	InflationLastTime        int64        // timestamp of last processing of inflation
	Inflation                rational.Rat // current annual inflation rate
    DateLastCommissionReset  int64        // unix timestamp for last commission accounting reset
    FeePool                  coin.Coins   // fee pool for all the fee shares which have already been distributed
    AccumAdjustment          rational.Rat // Adjustment factor for calculating global fee accum
}
```

### The Queue

The queue is ordered so the next to unbond/re-delegate is at the head. Every
tick the head of the queue is checked and if the unbonding period has passed
since `InitHeight` commence with final settlement of the unbonding and pop the
queue.

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

## Validator-Candidate

The `Candidate` struct holds the current state and some historical actions of
validators or candidate-validators. 

``` golang
type Candidate struct {
	Status                 CandidateStatus       
	PubKey                 crypto.PubKey
	Owner                  auth.Account
	GlobalStakeShares      rational.Rat 
	IssuedDelegatorShares  rational.Rat
	RedelegatingShares     rational.Rat
	VotingPower            rational.Rat 
    Commission             rational.Rat
    CommissionMax          rational.Rat
    CommissionChangeRate   rational.Rat
    CommissionChangeToday  rational.Rat
    ProposerRewardPool     coin.Coins
    AccumAdjustment        rational.Rat
    Description            Description 
}

type CandidateStatus byte
const (
    Bonded    CandidateStatus = 0x00
    Unbonding CandidateStatus = 0x01
    Unbonded  CandidateStatus = 0x02
)

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
 - PubKey: separated key from the owner of the candidate as is used strictly
   for participating in consensus.
 - Owner: Address where coins are bonded from and unbonded to 
 - GlobalStakeShares: Represents shares of `GlobalState.BondedPool` if
   `Candidate.Status` is `Bonded`; or shares of `GlobalState.UnbondedPool` if
   `Candidate.Status` is otherwise
 - IssuedDelegatorShares: Sum of all shares issued to delegators (which
   includes the candidate's self-bond) which represent each of their stake in
   the Candidate's `GlobalStakeShares`
 - RedelegatingShares: The portion of `IssuedDelegatorShares` which are
   currently re-delegating to a new validator
 - VotingPower: Proportional to the amount of bonded tokens which the validator
   has if the validator is within the top 100 validators.
 - Commission:  The commission rate of fees charged to any delegators
 - CommissionMax:  The maximum commission rate which this candidate can charge 
   each day from the date `GlobalState.DateLastCommissionReset` 
 - CommissionChangeRate: The maximum daily increase of the candidate commission
 - CommissionChangeToday: Counter for the amount of change to commission rate 
   which has occurred today, reset on the first block of each day (UTC time)
 - ProposerRewardPool: reward pool for extra fees collected when this candidate
   is the proposer of a block
 - AccumAdjustment: Adjustment factor for calculating fee accum for the
   candidate
 - Description
   - Name: moniker
   - DateBonded: date determined which the validator was bonded
   - Identity: optional field to provide a signature which verifies the
     validators identity (ex. UPort or Keybase)
   - Website: optional website link
   - Details: optional details

validator candidacy can be declared using the `TxDeclareCandidacy` transaction.
During this transaction a self-delegation transaction is executed to bond
tokens which are sent in with the transaction.

``` golang
type BondUpdate struct {
	PubKey crypto.PubKey
	Amount coin.Coin       
}

type TxDeclareCandidacy struct {
	BondUpdate
    Description         Description
	Commission          int64  
	CommissionMax       int64 
	CommissionMaxChange int64 
}
``` 

For all subsequent self-bonding, whether self-bonding or delegation the
`TxDelegate` function should be used. In this context `TxUnbond` is used to
unbond either delegation bonds or validator self-bonds. 

### Store indexing

Within the store, each `Candidate` is stored by validator-pubkey.

 - key: validator-pubkey
 - value: `ValidatorBond` object

A second key-value pair is also maintained by the store in order to quickly
sort though the group of all candidates: 

 - key: `Candidate.GlobalStakeShares`
 - value: `ValidatorBond.PubKey`

When the set of all validators needs to be determined from the group of all
candidates, the top 100 candidates, sorted by GlobalStakeShares can be retrieved
from this sorting without the need to retrieve the entire group of candidates.  

### Validators getting kicked

Unbonding of an entire validator-candidate to a temporary liquid account occurs
under the scenarios: 
 - the owner unbonds all of their staked tokens
 - validator liveliness issues
 - crosses a self-imposed safety threshold
   - minimum number of tokens staked by owner
   - minimum ratio of tokens staked by owner to delegator tokens 

When this occurs delegator's tokesn do not unbond to their personal wallets but
begin the unbonding process to a pool where they must then transact in order to
withdraw to their respective wallets. The following unbonding will use the
following queue element

``` golang
type QueueElemUnbondCandidate struct {
	QueueElem
}
```

If a delegator chooses to initiate an unbond or re-delegation of their shares
while a candidate-unbond is commencing, then that unbond/re-delegation is
subject to a reduced unbonding period based on how much time those funds have
already spent in the unbonding queue.

If the validator was kicked for liveliness issues and is able to regain
liveliness then all delegators in the temporary unbonding pool which have not
transacted to move will be bonded back to the now-live validator and begin to
once again collect provisions and rewards. 

## Delegator bond

Atom holders may delegate coins to validators, under this circumstance their
funds are held in a `DelegatorBond`. It is owned by one delegator, and is
associated with the shares for one validator. The sender of the transaction is
considered to be the owner of the bond,  

``` golang
type DelegatorBond struct {
	Candidate           crypto.PubKey
	Shares              rational.Rat
    FeeWithdrawalHeight int64  
} 
```

Description: 
 - Candidate: pubkey of the validator candidate bonding too
 - Shares: the number of shares received from the validator candidate
 - FeeWithdrawalHeight: last height fee rewards were withdrawn from
   `Candidate.FeeShares`

Each `DelegatorBond` is individually indexed within the store by delegator
address and candidate pubkey.

 - key: Delegator and Candidate-Pubkey
 - value: DelegatorBond 


### Delegating

Delegator bonds are created using the TxDelegate transaction. Within this
transaction the validator candidate queried with an amount of coins, whereby
given the current exchange rate of candidate's delegator-shares-to-atoms the
candidate will return shares which are assigned in `DelegatorBond.Shares`.

``` golang 
type TxDelegate struct { 
    BondUpdate 
}
```

### Unbonding

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


### Re-Delegation

``` golang
type QueueElemReDelegate struct {
	QueueElem
	Payout       auth.Account  // account to pay out to
    Shares       rational.Rat  // amount of shares which are unbonding
    NewCandidate crypto.PubKey // validator to bond to after unbond
}
```

The re-delegation command allows delegators to switch validators while still
receiving equal reward to as if you had never unbonded. Transactions structs of
the re-delegation command may looks like this:

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
the `Candidate.Shares` and the term `RedelegatingShares` is incremented.
During the unbonding period all unbonding shares do not count towards the
voting power of a validator. Once the `QueueElemReDelegation` has reached
maturity, the appropriate unbonding shares are removed from the `Shares` and
`RedelegatingShares` term.   


### Cancel Unbonding

A delegator who is in the process of unbonding from a validator may use the
re-delegate transaction to bond back to the original validator they're
currently unbonding from (and only that validator). If initiated, the delegator
will immediately begin to one again collect rewards from their validator. 


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
validators `GlobalStakeShares`.  This calculation ensures that the worth of the
`GlobalStakeShares` of other validators remains worth a constant absolute
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

Note that when an re-delegation occurs the shares to move are placed in an
re-delegation queue where they continue to collect validator provisions until
queue element matures. Although provisions are collected during re-delegation,
re-delegation tokens do not contribute to the voting power of a validator. 

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
which needs to be updated is the `BondedTokenPool`. So for each previsions
cycle:

```
params.BondedTokenPool += provisionTokensHourly
```

## Fee Calculations

Collected fees are pooled globally and divided out passively to validators and
delegators. Each validator has the opportunity to charge commission to the
delegators on the fees collected on behalf of the delegators by the validators.
Fees are paid directly into a global fee pool. With regards to fee withdrawal: 
 
 - when withdrawing one must withdrawal the maximum amount they are entitled
   too, leaving nothing in the pool, 
 - when bonding, unbonding, or re-delegating tokens to an existing account a
   full withdrawal of the fees must occur (as the rules for lazy accounting
   change).

When the validator is the proposer of the round, that validator (and their
delegators) receives between 1% and 5% of fee rewards, the remainder is
distributed socially by voting power to all validators including the proposer
validator.  The amount of provision reward is calculated from pre-commits
Tendermint messages. All provision rewards are added to a provision reward pool 
which validator holds individually. Here note that `BondedShares` represents
the sum of all voting power saved in the `GlobalState` (denoted `gs`).

```
proposerReward = feesCollected * (0.01 + 0.04 
                  * sumOfVotingPowerOfPrecommitValidators / gs.BondedShares)
candidate.ProposerRewardPool += proposerReward

distributedReward = feesCollected - proposerReward
gs.FeePool += distributedReward
```

The entitlement to the fee pool held by the each validator is accounted for
lazily.  First we must account for a candidate's `count` and `adjustment`. The
`count` represents a lazy accounting of what that candidates entitlement to the
fee pool would be if there `VotingPower` was to never change and they were to
never withdraw fees. 

``` 
candidate.count = candidate.VotingPower * BlockHeight
``` 

Similarly the GlobalState count can be passively calculated whenever needed,
where `BondedShares` is the updated sum of voting powers from all validators.

``` 
gs.count = gs.BondedShares * BlockHeight
``` 

The `adjustment` term accounts for changes in voting power and withdrawals of
fees. The adjustment factor must be persisted with the candidate and modified
whenever fees are withdrawn from the candidate or the voting power of the
candidate changes. When the voting power of the candidate changes the
`AccumAdjustment` factor is increased/decreased by the cumulative difference in the
voting power if the voting power has been the new voting power as opposed to
the old voting power for the entire duration of the blockchain up the previous
block. Each time there is an adjustment change the GlobalState (denoted `gs`)
`AccumAdjustment` must also be updated.

```
AdjustmentChange = candidate.VotingPower * (height - 1) 
                   - candidate.PreviousBlockVotingPower * (height - 1)
candidate.AccumAdjustment += AdjustmentChange
gs.AccumAdjustment += AdjustmentChange
```

Note that this means that the adjustment factor can be negative if the voting
power of a candidate is decreasing. When fees are withdrawn the adjustment
factor is reduced to account for those missing fees, by the `accum`
(calculation to determine `accum` for  transaction described later) 

``` 
candidate.AccumAdjustment += accum
gs.AccumAdjustment += accum
``` 

With each candidates and the global state  `count` and `AccumAdjustment` known the
entitlement to the global fee pool can be calculated for any candidate.

```
candidate.accum = candidate.count - candidate.AccumAdjustment
gs.accum = gs.count - gs.AccumAdjustment
candidate.entitlement = candidate.accum / gs.accum
candidate.feeRewards = candidate.entitlement * gs.FeePool
```

So far we have covered two sources fees which can be withdrawn from: Fees from
proposer rewards (`candidate.ProposerRewardPool`), and fees from the fee pool
(`candidate.feeRewards`). However we should note that all fees from fee pool
are subject to commission rate from the owner of the candidate. These next
calculations outline the math behind withdrawing fee rewards as either a
delegator to a candidate or as the owner of a candidate.

### As a delegator

As a delegator each time you withdraw fees you must draw all fees which you are
entitled too.  This is achieved by calculating an accum based on your share of
the candidate as well as the duration of time since your last withdrawal. Here
`bond` refers to a `DelegatorBond` object. 

```
accumRaw = candidate.VotingPower * bond.Shares / Candidate.IssuedDelegatorShares 
         * (BlockHeight - bond.FeeWithdrawalHeight)
``` 

Accum must be further reduced as there is a commission rate which is being
charged by the candidate, therefor the actual accum which is being withdrawn
from is reduced

```
accum = accumRaw * (1 - candidate.Commission)
```

With `accum` solved the amount of entitled fees from both the
`ProposerRewardPool` and the `feeRewards` can be determined 

```
rewardProposer = candidate.ProposerRewardPool * accum / candidate.accum
rewardPool = candidate.feeRewards * accum / candidate.accum
```

When fees are withdrawn, the `candidate.AccumAdjustment` must be increased, and
account balances must be updated:

```
candidate.AccumAdjustment += accum
candidate.ProposerRewardPool -= rewardProposer
gs.FeePool -= rewardPool
delegator.Owner.Balance += rewardProposer + rewardPool
```

As mentioned earlier, every time the voting power of a delegator bond is
changing either by unbonding or further bonding, all fees must be first
withdrawn. 

### As the candidate owner

As the candidate owner the same calculations apply as if you were a delegator
however  you are also entitled to a commission on the fee rewards received by
all delegators. The `accumComm` term is calculated similarly to the `accum`
term however no ratio of bond shares to total shares must be multiplied as the
commission applies to the entire stake of the candidate. 

```
accumComm = candidate.VotingPower * (BlockHeight - bond.FeeWithdrawalHeight) 
            * candidate.Commission

rewardProposerComm = candidate.ProposerRewardPool * accumComm / candidate.accum
rewardPoolComm = candidate.feeRewards * accumComm / candidate.accum
``` 

Upon withdrawal as a candidate owner the final accounting is calculated as so:

```
candidate.AccumAdjustment += accum + accumComm
candidate.ProposerRewardPool -= rewardProposer + rewardProposerComm
gs.FeePool -= rewardPool + rewardPoolComm
delegator.Owner.Balance += rewardProposer + rewardPool
                           + rewardProposerComm + rewardPoolComm
```

### Other general notes on fees accounting

- When a delegator chooses to re-delegate shares, fees continue to accumulate
  until the re-delegation queue reaches maturity. At the block which the queue
  reaches maturity and shares are re-delegated all available fees are
  simultaneously withdrawn. 
- Whenever a totally new validator is added to the validator set, the `accum`
  of the entire candidate must be 0, meaning that the initial value for
  `candidate.AccumAdjustment` must be set to the value of `canidate.Count` for the
  height which the candidate is added on the validator set.
- The accum of a new delegator bond will be 0 for the height at which the bond
  was added. This is achieved by setting `DelegatorBond.FeeWithdrawalHeight` to
  the height which the bond was added. 
