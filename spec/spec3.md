# Staking Module

### Basic Terms and Definitions

- Cosmsos Hub - a Tendermint-based Proof of Stake blockchain system
- Atom - native token of the Cosmsos Hub
- Atom holder - an entity that holds some amount of Atoms 
- Candidate - an Atom holder that is actively involved in Tendermint blockchain protocol (running Tendermint Full Node
TODO: add link to Full Node definition) and is competing with other candidates to be elected as a Validator 
(TODO: add link to Validator definition))
- Validator - a candidate that is currently selected among a set of candidates to be able to sign protocol messages 
in the Tendermint consensus protocol 
- Delegator - an Atom holder that has bonded any of its Atoms by delegating them to a validator (or a candidate) 
- Bonding Atoms - a process of locking Atoms in a bond deposit (putting Atoms under protocol control).
Atoms are always bonded through a validator (or candidate) process. Bonded atoms can be slashed (burned) in case a 
validator process misbehaves (does not behave according to a protocol specification). Atom holder can regain access 
to it's bonded Atoms (if they are not slashed in the meantime), i.e., they can be moved to it's account, 
after Unbonding period has expired. 
- Unbonding period - a period of time after which Atom holder gains access to its bonded Atoms (they can be withdrawn 
to a user account) or they can re-delegate  
- Inflationary provisions - inflation is a process of increasing Atom supply. Atoms are being created in the process of 
(Cosmos Hub) blocks creation. Owners of bonded atoms are rewarded for securing network with inflationary provisions 
proportional to it's bonded Atom share.
- Transaction fees - transaction fee is a fee that is included in the Cosmsos Hub transactions. The fees are collected 
by the current validator set and distributed among validators and delegators in proportion to it's bonded Atom share.       
- Commission fee - a fee taken from the transaction fees by a validator for it's service 

### Overview

The Cosmos Hub is a Tendermint-based Proof of Stake blockchain system that serves as a backbone of the Cosmos ecosystem.
It is operated and secured by an open and globally decentralized set of validators. Tendermint consensus is a 
Byzantine fault-tolerant distributed protocol that involves all validators in the process of exchanging protocol 
messages in the production of each block. To avoid Nothing-at-Stake problem, a validator in Tendermint needs to lock up
coins in a bond deposit. Tendermint protocol messages are signed by the validator's private key, and this is a basis for 
Tendermint strict accountability that allows punishing misbehaving validators by slashing (burning) their bonded Atoms. 
On the other hand, validators are for it's service of securing blockchain network rewarded by the inflationary 
provisions and transactions fees. This incentivizes correct behavior of the validators and provide economic security 
of the network.

The native token of the Cosmos Hub is called Atom; becoming a validator of the Cosmos Hub requires holding Atoms. 
However, not all Atom holders are validators of the Cosmos Hub. More precisely, there is a selection process that 
determines the validator set as a subset of all validator candidates (Atom holder that wants to 
become a validator). The other option for Atom holder is to delegate their atoms to validators, i.e., 
being a delegator. A delegator is an Atom holder that has bonded its Atoms by delegating it to a validator 
(or validator candidate). By bonding Atoms to securing network (and taking a risk of being slashed in case the 
validator misbehaves), a user is rewarded with inflationary provisions and transaction fees proportional to the amount 
of its bonded Atoms. The Cosmos Hub is designed to efficiently facilitate a small numbers of validators (hundreds), and 
large numbers of delegators (tens of thousands). More precisely, it is the role of the Staking module of the Cosmos Hub 
to support various staking functionality; including validator set selection, delegating, bonding and withdrawing Atoms, 
and the distribution of inflationary provisions and transaction fees.

### The pool and the share

At the core of the Staking module is the concept of a pool which denotes collection of Atoms contributed by different 
Atom holders. There are two global pools in the Staking module: bonded pool and unbonded pool. Bonded Atoms are part 
of the global bonded pool. On the other side, if a candidate or delegator wants to unbond its Atoms, those Atoms are 
kept in the unbonding pool for a duration of the unbonding period. In the Staking module, a pool is logical concept, 
i.e., there is no pool data structure that would be responsible for managing pool resources. Instead, it is managed 
in a distributed way. More precisely, at the global level, for each pool, we track only the total amount of 
(bonded or unbonded) Atoms and a current amount of issued shares. A share is a unit of Atom distribution and the 
value of the share (share-to-atom exchange rate) is changing during the system execution. The 
share-to-atom exchange rate can be computed as:

`share-to-atom-ex-rate = size of the pool / ammount of issued shares`

Then for each candidate (in a per candidate data structure) we keep track of an amount of shares the candidate is owning 
in a pool. At any point in time, the exact amount of Atoms a candidate has in the pool
can be computed as the number of shares it owns multiplied with the share-to-atom exchange rate:

`candidate-coins = candidate.Shares * share-to-atom-ex-rate`

The benefit of such accounting of the pool resources is the fact that a modification to the pool because of 
bonding/unbonding/slashing/provisioning of atoms affects only global data (size of the pool and the number of shares) 
and the related validator/candidate data structure, i.e., the data structure of other validators do not need to be 
modified. Lets explain this further with several small examples: 

We consider initially 4 validators p1, p2, p3 and p4, and that each validator has bonded 10 Atoms 
to a bonded pool. Furthermore, lets assume that we have issued initially 40 shares (note that the initial distribution 
of the shares, i.e., share-to-atom exchange rate can be set to any meaningful value), i.e., 
share-to-atom-ex-rate = 1 atom per share. Then at the global pool level we have, the size of the pool is 40 Atoms, and
the amount of issued shares is equal to 40. And for each validator we store in their corresponding data structure
that each has 10 shares of the bonded pool. Now lets assume that the validator p4 starts process of unbonding of 5 
shares. Then the total size of the pool is decreased and now it will be 35 shares and the amount of Atoms is 35. 
Note that the only change in other data structures needed is reducing the number of shares for a validator p4 from 10 
to 5.

Lets consider now the case where a validator p1 wants to bond 15 more atoms to the pool. Now the size of the pool
is 50, and as the exchange rate hasn't changed (1 share is still worth 1 Atom), we need to create more shares, 
i.e. we now have 50 shares in the pool in total.
Validators p2, p3 and p4 still have (correspondingly) 10, 10 and 5 shares each worth of 1 atom per share, so we 
don't need to modify anything in their corresponding data structures. But p1 now has 25 shares, so we update the amount
of shares owned by the p1 in its data structure. Note that apart from the size of the pool that is in Atoms, all other
data structures refer only to shares.

Finally, lets consider what happens when new Atoms are created and added to the pool due to inflation. Lets assume that 
the inflation rate is 10 percent and that it is applied to the current state of the pool. This means that 5 Atoms are 
created and added to the pool and that each validator now proportionally increase it's Atom count. Lets analyse how this 
change is reflected in the data structures. First, the size of the pool is increased and is now 55 atoms. As a share of 
each validator in the pool hasn't changed, this means that the total number of shares stay the same (50) and that the 
amount of shares of each validator stays the same (correspondingly 25, 10, 10, 5). But the exchange rate has changed and 
each share is now worth 55/50 Atoms per share, so each validator has effectively increased amount of Atoms it has. 
So validators now have (correspondingly) 55/2, 55/5, 55/5 and 55/10 Atoms. 

The concepts of the pool and its shares is at the core of the accounting in the Staking module. It is used for managing 
the global pools (such as bonding and unbonding pool), but also for distribution of Atoms between validator and its 
delegators (we will explain this in section X).       
  
### State

The staking module persists the following information to the store:
- `GlobalState`, describing the global pools
- a `Candidate` for each candidate validator, indexed by public key
- a `Candidate` for each candidate validator, indexed by shares in the global pool (ie. ordered)
- a `DelegatorBond` for each delegation to a candidate by a delegator, indexed by delegator and candidate
    public keys
- a `Queue` of unbonding delegations (TODO)

## Global State

GlobalState data structure is responsible for managing total Atom supply, and two global Atom pools: bonded pool and 
unbonded pool. In addition to the global Atoms pools, GlobalState 
also keep information about the current annual inflation rate, and the timestamp of the last inflation processing.
TODO: Add description of the other fields.    

``` golang
type GlobalState struct {
	TotalSupply              int64        // total supply of Atoms
	BondedShares             rational.Rat // sum of all shares distributed for the BondedPool
	UnbondedShares           rational.Rat // sum of all shares distributed for the UnbondedPool
	BondedPool               int64        // reserve of bonded tokens
	UnbondedPool             int64        // reserve of unbonded tokens held with candidates
	InflationLastTime        int64        // timestamp of last processing of inflation
	Inflation                rational.Rat // current annual inflation rate
    DateLastCommissionReset  int64        // unix timestamp for last commission accounting reset
    FeePool                  coin.Coins   // fee pool for all the fee shares which have already been distributed
    ReservePool              coin.Coins   // pool of reserve taxes collected on all fees for governance use
    Adjustment               rational.Rat // Adjustment factor for calculating global fee accum
}
```

### Candidate

The `Candidate` data structure holds the current state and some historical actions of
validators or candidate-validators. 

``` golang
type Candidate struct {
	Status                 CandidateStatus       
	PubKey                 crypto.PubKey
	GovernancePubKey       crypto.PubKey
	Owner                  Address
	GlobalStakeShares      rational.Rat 
	IssuedDelegatorShares  rational.Rat
	RedelegatingShares     rational.Rat
	VotingPower            rational.Rat 
    Commission             rational.Rat
    CommissionMax          rational.Rat
    CommissionChangeRate   rational.Rat
    CommissionChangeToday  rational.Rat
    ProposerRewardPool     coin.Coins
    Adjustment             rational.Rat
    Description            Description 
}

type CandidateStatus byte
const (
    VyingUnbonded  CandidateStatus = 0x00
    VyingUnbonding CandidateStatus = 0x01
    Bonded         CandidateStatus = 0x02
    KickUnbonding  CandidateStatus = 0x03
    KickUnbonded   CandidateStatus = 0x04
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
 - Status: signal that the candidate is either vying for validator status
   either unbonded or unbonding, an active validator, or a kicked validator
   either unbonding or unbonded.
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
 - Adjustment factor used to passively calculate each validators entitled fees
   from `GlobalState.FeePool`
 - Description
   - Name: moniker
   - DateBonded: date determined which the validator was bonded
   - Identity: optional field to provide a signature which verifies the
     validators identity (ex. UPort or Keybase)
   - Website: optional website link
   - Details: optional details


Candidates are indexed by their `Candidate.PubKey`.
Additionally, we index empty values by the candidates global stake shares concatenated with the public key.

TODO: be more precise.

When the set of all validators needs to be determined from the group of all
candidates, the top candidates, sorted by GlobalStakeShares can be retrieved
from this sorting without the need to retrieve the entire group of candidates.
When validators are kicked from the validator set they are removed from this
list. 

### TxDelegate

All bonding, whether self-bonding or delegation, is done via
`TxDelegate`. 

Delegator bonds are created using the TxDelegate transaction. Within this
transaction the validator candidate queried with an amount of coins, whereby
given the current exchange rate of candidate's delegator-shares-to-atoms the
candidate will return shares which are assigned in `DelegatorBond.Shares`.

``` golang 
type TxDelegate struct { 
	PubKey crypto.PubKey
	Amount coin.Coin       
}
```

### DelegatorBond

Atom holders may delegate coins to validators, under this circumstance their
funds are held in a `DelegatorBond`. It is owned by one delegator, and is
associated with the shares for one validator. The sender of the transaction is
considered to be the owner of the bond,  

``` golang
type DelegatorBond struct {
	Candidate            crypto.PubKey
	Shares               rational.Rat
    AdjustmentFeePool    coin.Coins  
    AdjustmentRewardPool coin.Coins  
} 
```

Description: 
 - Candidate: pubkey of the validator candidate: bonding too
 - Shares: the number of shares received from the validator candidate
 - AdjustmentFeePool: Adjustment factor used to passively calculate each bonds
   entitled fees from `GlobalState.FeePool`
 - AdjustmentRewardPool: Adjustment factor used to passively calculate each
   bonds entitled fees from `Candidate.ProposerRewardPool``

Each `DelegatorBond` is individually indexed within the store by delegator
address and candidate pubkey.

 - key: Delegator and Candidate-Pubkey
 - value: DelegatorBond 


## Provision Calculations

Every hour atom provisions are assigned proportionally to the each slashable
bonded token which includes re-delegating atoms but not unbonding tokens.

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



