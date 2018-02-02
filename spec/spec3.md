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
It is secured by an open and globally decentralized set of validators. Tendermint consensus is a Byzantine fault-tolerant 
distributed protocol that involves all validators in the process of exchanging protocol messages in the production of 
each block. To avoid Nothing-at-Stake problem, a validator in Tendermint needs to lock up coins in a bond 
deposit. Tendermint protocol messages are signed by the validator's private key, and this is a basis for Tendermint 
strict accountability that allows punishing misbehaving validators by slashing (burning) their bonded Atoms. On the 
other hand, validators are for it's service of securing blockchain network rewarded by the inflationary provisions and 
transactions fees. This incentivizes correct behavior of the validators and provide economic security of the network.

The native token of the Cosmos Hub is called Atom; becoming a validator of the Cosmos Hub requires holding Atoms. 
However, not all Atom holders are validators of the Cosmos Hub. More precisely, there is a selection process, that is 
regularly executed, that determines the validator set as a subset of all validator candidates (Atom holder that wants to 
become a validator). The other option for Atom holder is to delegate their atoms to other validators, i.e., 
being a delegator. A delegator is an Atom holder that has bonded its Atoms by delegating it to a validator 
(or validator candidate). By contributing Atoms to securing network (and taking a risk of being slashed in case the 
validator misbehaves), a user is rewarded with inflationary provisions and transaction fees proportional to the amount 
of its bonded Atoms. The Cosmos Hub is designed to efficiently facilitate small numbers of validators (hundreds), and 
large numbers of delegators (tens of thousands). More precisely, it is the role of the Staking module of the Cosmos Hub 
to support various staking functionality; including validator set selection, delegating, bonding and withdrawing Atoms, 
and the distribution of inflationary provisions and transaction fees.

### State

The staking module persists the following information to the store:
- `GlobalState`, describing the global pools
- a `Candidate` for each candidate validator, indexed by public key
- a `Candidate` for each candidate validator, indexed by shares in the global pool (ie. ordered)
- a `DelegatorBond` for each delegation to a candidate by a delegator, indexed by delegator and candidate
    public keys
- a `Queue` of unbonding delegations (TODO)

## Global State

There are two global Atom pools managed by the Staking module: bonded pool and unbonded pool. When an Atom holder bonds
some Atoms they becomes part of the global bonded pool.  

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

The `Candidate` struct holds the current state and some historical actions of
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



