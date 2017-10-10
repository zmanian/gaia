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

### Local-Test Example

Here is a quick example to get you off your feet: 

First generate a new key, save your name and pubkey for later

```
gaiacli keys new 
gaiacli keys list 
<generate your new key under YOURNAME, save YOURPUBKEY>
```

Next initialize a gaia chain and your account a bunch of fun fake money

```
gaia init <YOURPUBKEY> --home=$HOME/.atlas1 --chain-id=test 
```

Also initialize a second chain which will be used to connect to to represent 
your validaator. Copy the genesis from the first chain in.

```
gaia init <YOURPUBKEY> --home=$HOME/.atlas2 --chain-id=test
cp $HOME/.atlas1/genesis.json $HOME/.atlas2/genesis.json 
```

Additionally open the file `$HOME/.atlas2/config.toml` and modify it 
to use different addresses from the first node as well as connect to the 
first node. Your config should look as follows:

```
proxy_app = "tcp://127.0.0.1:46668"
moniker = "anonymous"
fast_sync = true
db_backend = "leveldb"
log_level = "state:info,*:error"

[rpc]
laddr = "tcp://0.0.0.0:46667"

[p2p]
laddr = "tcp://0.0.0.0:46666"
seeds = "tcp://0.0.0.0:46656"
```

Great, now that we've initialized the chains, let's 
start the first and second node.
```
gaia start --home=$HOME/atlas1
gaia start --home=$HOME/atlas2
```

In a separate terminal window initialize the client to the first node

```
gaiacli init --chain-id=test --node=tcp://localhost:46657
```

Now bond some coins and check out the validator set. You should see that coins
have moved from your account balance and to the validator account. 
The validator pubkey must be registered using the `--pubkey` flag. 
In this example, your validator pubkey which you have registered can be found in:
`~/.atlas2/priv_validator.json`

```
gaiacli query account <YOURPUBKEY>
gaiacli tx bond --amount 2strings --name <YOURNAME> --pubkey <VALIDATORPUBKEY>
gaiacli query validators
gaiacli query account <YOURPUBKEY>
``` 

After your coins have been bonded you should start to see blocks rolling in on the 
second node. If the second node is halted then the first node should continue publishing blocks
as it has a voting power of 10 (from genesis.json) which is greater than the number 
of tokens we bonded in the previous step. If we had bonded a larger amount such as 11 _strings_ 
then the network would have halted when to paused the second node. Give it a shot!

Finally watch all your power relinquish as you unbond some coins, you should see your
VotingPower reduce and your account balance increase.

```
gaiacli tx unbond --amount 1strings --name <YOURNAME>
gaiacli query validators
gaiacli query account <YOURPUBKEY>
``` 
