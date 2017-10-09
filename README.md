# Cosmos-Hub Basic Staking

This project is a demonstration of the Cosmos-Hub with basic staking
functionality staking module designed to get validators acquainted
with staking concepts and procedures.

Currently, the validator set is updated every block. The validator set is
determined as the validators with the top 100 bonded atoms. Currently, all
bonding and unbonding is instantaneous (no queue). Absent features include,
delegation, validator rewards, unbonding wait period.

### Installation
```
go get github.com/cosmos/gaia 
cd $GOPATH/src/github.com/cosmos/gaia
make all
```

### Example

Here is a quick example to get you off your feet: 

First generate a new key, save your name and pubkey for later

```
gaiacli keys new 
gaiacli keys list 
<generate your new key under YOURNAME, save YOURPUBKEY>
```

Next initialize a gaia chain and your account a bunch of fun fake money

```
gaia init --chain-id=test <YOURPUBKEY>
gaia start
```

In a separate terminal window initialize the client

```
gaiacli init --chain-id=test --node=tcp://localhost:46657
```

Now bond some coins and check out the validator set. You should see that coins
have moved from your account balance and to the validator account. If the validator
pubkey is the same as the account who is sending the coins then the the `--pubkey`
flag does not need to be used. Otherwise the validator can have a pubkey registered
using the `--pubkey` flag. In this example, this pubkey can be found within 
`~/.cosmos-gaia/priv_validator.json`

```
gaiacli query account <YOURPUBKEY>
gaiacli tx bond --amount 2strings --name <YOURNAME> --pubkey <VALIDATORPUBKEY>
gaiacli query validators
gaiacli query account <YOURPUBKEY>
``` 

Next unbond some coins, you should see your VotingPower reduce 
and your account balance increase.

```
gaiacli tx unbond --amount 1strings --name <YOURNAME>
gaiacli query validators
gaiacli query account <YOURPUBKEY>
``` 
