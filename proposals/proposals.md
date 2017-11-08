# Staking Proposals

This document is intended to lay out the implementation order of various
staking features. Each iteration of the staking module should be deployable 
as a test-net. Please contribute to this document and add discussion :)

Overall the Staking module should be rolled out in the following steps

1. Self-Bonding
2. Delegation
4. Unbonding period
5. Rewards, Commission, Fee Pool
6. Delegator Rebonding

## Self-Bonding

The majority of the processing will happen from the set of all Validator Bonds
which is defined as an array and persisted in the application state.  Each
validator bond is defined as the object below. 
 - Pubkey: Validator PubKey
 - Owner: Account which coins are bonded from and unbonded to
 - BondedTokens: Total number of bond tokens for the validator
 - VotingPower: Voting power in consensus

``` golang
type ValidatorBonds []*ValidatorBond

type ValidatorBond struct {
	PubKey       crypto.PubKey 
	Owner        sdk.Actor    
	BondedTokens uint64    
	VotingPower  uint64   
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

The next phase of development includes delegation functionality. All transactions
are still instantaneous in this development level.

The previous struct `ValidatorBond` is now split up into two structs which represent
the candidate account or the bond to a candidate account.

Each validator-candidate bond is defined as the object below. 
 - Pubkey: Candidate PubKey
 - Owner: Account which coins are bonded from and unbonded to
 - Shares: Total number of shares issued by this candidate
 - VotingPower: Voting power in consensus

``` golang
type Candidate struct {
	PubKey       crypto.PubKey
	Owner        sdk.Actor
	Shares       uint64    
	VotingPower  uint64   
}
```

At this phase in development the number of shares issued to each account 
will be equal to the number of coins bonded to this candidate. In later phases
when staking reward is introduced, the shares will not be equal to the number 
of coins bonded to the candidate. 


DelegatorBond represents some bond tokens held by an account. It is owned by
one delegator, and is associated with the voting power of one delegatee.

``` golang
type DelegatorBond struct {
	PubKey crypto.PubKey
	Shares uint64
} 
```
The sender of the transaction is considered to be the owner of the bond being
sent.  In this backend the client is expected to keep track of the list of
pubkeys which the delegator has delegated to. 

The transactions types must be renamed. `TxBond` is now effectively
`TxDeclareCandidacy` and but is only used to become a new validator. For all
subsiquent bonding, whether self-bonding or delegation the `TxDelegate`
function should be used. In this context `TxUnbond` is used to unbond either
delegation bonds or validator self-bonds. As such `TxUnbond` must now include
the PubKey which the unbonding is to occur from.  

``` golang
type BondUpdate struct {
	PubKey crypto.PubKey
	Amount coin.Coin       
}

type TxDelegate struct { BondUpdate }
type TxDeclareCandidacy struct { BondUpdate }
type TxUnbond struct { BondUpdate }
```

## Unbonding Period

A staking-module state parameter will be introduced which defines the number of
blocks to unbond from a validator. None of the structs should be affected by
this phase. Once complete unbonding as a validator or a Candidate will put your
coins in a queue before being returned.

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
The queue element struct for unbonding should look like this: 

``` golang
type QueueElemUnbond struct {
	Candidate    crypto.PubKey
	Payout       sdk.Actor // account to pay out to
	HeightAtInit uint64 // when the queue was initiated
	BondShares   uint64    // amount of bond tokens which are unbonding
}
```

## Rewards, Commission, Fee Pool

Collected fees are payed directly to a global hold account and reflected in the
`HoldCoins` during each reward cycle.

Each validator will now have the opportunity to charge commission on the fees
collected which are to be forwarded to their delegators to their delegators.
Validators must specify the rate and the maximum value of the commission
charged to delegators as an act of self-regualtion.  Although the rate

In addition, within this development phase the total supply of atoms increases
with each reward cycle, however commisions is not charged to the atoms supply 
which is provided by the
 
``` golang
type Candidate struct {
	    PubKey               crypto.PubKey
	    Owner                sdk.Actor 
    	Shares               uint64    
    	HoldCoins            uint64  
    	VotingPower          uint64   
    	Commission           uint64
    	CommissionMax        uint64
    	CommissionChangeRate uint64
        FeeAccum             uint64
}
```

Several new parameters are introduced:
 - HoldCoins: Total number of coins held by this validator, distinct from
   shares, this term is used 
 - Commission: The current commission rate currently being charged by the
   validator
 - Commission:  The commission percent of fees charged to any delegators
 - CommissionMax:  The maximum commission rate which this validator can charge
 - CommissionChangeRate: The maximum change per reward cycle which the validator
   can change their commission by
 - FeeAccum: Cumulative counter for the amount of fees the candidate and its
   deligators are entitled too

With this model the exchange rate can be calculated as `HoldCoins/Shares`.  It
then follows that when unbonding coins, the coins to withdraw per share can be
calculated as the following:

```
unbondCoins = unbondShares / candidate.Shares * HoldCoins
```

`TxDeclareCandidacy` should be updated to include new relavent terms:

``` golang
type TxDeclareCandidacy struct {
	BondUpdate
	Commission          uint64  
	CommissionMax       uint64 
	CommissionMaxChange uint64 
}
```

During each reward cycle the amount of validator rewards will remain costant
by the delegators will remain constant during this process. The change in
validator self-delegated shares can be calculated solving from some fundamental
equations. 

In addition to the validator shares which each delegator holds, they also hold
a fee pool account which increases based on the amount of time or number of
blocks which have passed since the last withdrawal.

 - The piggy bank account can only be emptied out entirely every time it is
   withdrawn from 
 - Cannot add to an existing delegation until the piggy bank is emptied - maybe
   it is automatically emptied each time there is a delegation (if that is
   possible)
 - Piggy bank calculated on withdraw based on the delegation "shares" as well
   as block height or timestamp 

Depending, DelegatorBond may now look as follows: 

``` golang
type DelegatorBond struct {
	Candidate      crypto.PubKey
	Shares         uint64
    FeeAccumHeight uint64
    FeeAccum       uint64
} 
```

Additionally some new terms will need to be added into the Candidate:

``` golang
type Candidate struct {
	PubKey       crypto.PubKey
	Owner        sdk.Actor
	Shares       uint64    
	VotingPower  uint64   
    FeePool      []coin.Coin
    FeeAccum     uint64
}
```

Here a separate `PiggyBank` exists per candidate for every fee asset held by 
a validator. The `SumOfHeights` increments each time the rewards are given

```
PiggyBank.SumOfHeights = PiggyBank.SumOfHeights + 
                         (CurrentHeight - LastFeeRewardHeight)
```

For each fee token the calculation for the amount of withdrawal from the piggy
bank is: 

```
withdrawal = PiggyBank.FeeAsset.Amount 
            * (CurrentHeight - LastFeeWithdrawlHeight)
            / PiggyBank.SumOfHeights
```

For each withdrawal we must also remember to reduce the `SumOfHeights`

```
PiggyBank.SumOfHeights = PiggyBank.SumOfHeights - 
                         (CurrentHeight - LastFeeWithdrawlHeight)
```

## Delegator Rebonding

If a delegator is attempting to switch validators, using an unbonding command
followed by a bond command will reduce the amount of validation reward than
having just stayed with the same validator. Because we do not want to
disincentive delegators from switching validators, we may introduce a rebond
command which provides equal reward to as if you had never unbonded. 

The moment you are “rebonding” the old validator loses your voting power, and
the new validator gains this voting power. For the duration of the unbonding
period any rebonding account will be slashable by both the old validator and
the new validator. 

The mechanism for double slashability is as follows. When re-delegating the
hold account atoms are automatically moved from the old validator to the new
validator. Within the new validator - the atoms are treated as a regular
delegation. Within the old validator, the atoms have already been removed its
hold account but a separate mechanism applied to the re-delegation queue item -
in the rebonding queue - the fraction of all historical slashings on that
validator are recorded. When this queue reaches maturity if that total
slashings applied is greater on the old validator then the difference (amount
that should have been slashed from the first validator) is assigned as debt to
the delegator addresses. If a delegator address has dept - the only operation
they are permitted to perform is an unbond operation, additionally they cannot
vote, or accrue and new fees from the time the debt was added.

Transactions structs of the rebond command may looks like this: 

``` golang 
type TxRebond struct { 
    BondUpdate 
    Rebond crypto.PubKey 
} 
```

Where the `Rebond` is the new validator to rebond to. 

Within the delegation bond we will need to now add a debt as explained above:

``` golang
type DelegatorBond struct {
	PubKey           crypto.PubKey
	Shares           uint64
	Debt             uint64
    LastFeeWithdrawl time.Time
}
```
