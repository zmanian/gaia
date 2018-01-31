# Staking Module

## Basic Terms and Definitions

- Cosmsos Hub - a Tendermint-based Proof of Stake blockchain system
- Atom - native token of the Cosmsos Hub
- Atom holder - an entity that holds some amount of Atoms 
- Candidate - an Atom holder that is actively involved in Tendermint blockchain protocol (running Tendermint Full Node
TODO: add link to Full Node definition) and is competing with other candidates to be elected as a Validator 
(TODO: add link to Validator definition))
- Validator - a candidate that is currently selected among a set of candidates to be able to sign protocol messages 
in the Tendermint consensus protocol 
- Delegator - an Atom holder that has bonded subset of its Atoms by delegating them to a validator (or a candidate) 
- Bonding Atoms - a process of locking Atoms in a bond deposit (putting Atoms under protocol control).
Atoms are always bonded through a validator (or candidate) process. Bonded atoms can be slashed (burned) in case a 
validator process misbehaves (does not behave according to a protocol specification). Atom holder can regain access 
to it's bonded Atoms (if they are not slashed in the meantime), i.e., they can be moved to it's account, 
after Unbonding period has expired. 
- Unbonding period - a period of time after which Atom holder gains access to its bonded Atoms (they can be withdrawn 
to a user account) 
- Inflationary provisions - inflation is a process of increasing Atom supply. Atoms are being created in the process of 
(Cosmos Hub) blocks creation. Owners of bonded atoms are rewarded for securing network with inflationary provisions 
proportional to it's bonded Atom share.
- Transaction fees - transaction fee is a fee that is included in the Cosmsos Hub transactions. The fees are collected 
by the current validator set and distributed among validators and delegators in proportion to it's bonded Atom share.       
- Commission fee - a fee taken from the transaction fees by a validator for it's service 

## Overview

The Cosmos Hub is a Tendermint-based Proof of Stake blockchain system that serves as a backbone of the Cosmos ecosystem.
It is secured by an open and globally decentralized set of validators. Tendermint consensus is a Byzantine fault-tolerant 
distributed protocol that involves all validators in the process of exchanging protocol messages in the production of 
each block. To avoid Nothing-at-Stake problem, a validator in Tendermint needs to lock some of its coins in a bond 
deposit. Tendermint protocol messages are signed by the validator's private key, and this is a basis for Tendermint 
strict accountability that allows punishing misbehaving validators by slashing (burning) their bonded Atoms. On the 
other hand, validators are for it's service of securing blockchain network rewarded by the inflationary provisions and 
transactions fees. This incentivize correct behavior of the validators and provide economic security of the network.

The native token of the Cosmos Hub is called Atom, so becoming a validator of the Cosmos Hub requires to be Atom holder. 
However, not all Atom holders are validators of the Cosmos Hub. More precisely, there is a selection process, that is 
regularly executed, that selects the validator set as a subset of all validator candidates (Atom holder that wants to 
become a validator). Becoming a validator is not the only way for Atom holder to be involved in securing network 
and obtaining benefits for this service. The other option is becoming a delegator. A delegator is an Atom holder that 
has bonded a subset of its Atoms by delegating it to a validator (or validator candidate). By contributing Atoms to 
securing network (and taking a risk of being slashed in case the validator misbehaves), a user is rewarded with 
inflationary provisions and transaction fees proportional to the amount of its bonded Atoms. Cosmos Hub is designed to 
efficiently facilitate small numbers of validators (hundreds), and large numbers of delegators (tens of thousands). 
More precisely, it is the role of the Staking module of the Cosmos Hub to support various staking functionality,
including validator set selection, delegating, bonding and withdrawing Atoms, and the distribution of inflationary 
provisions and transaction fees.




