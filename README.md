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

### Atlas Test-Net Example

To work on the Atlas you will need some tokens to get started. 
To do this first generate a new key: 

```
MYNAME=<your name>
gaiacli keys new $MYNAME
gaiacli keys list
MYADDR=<your newly generated address>
```

Then get a hold of us through [Riot](https://riot.im/app/#/room/#cosmos:matrix.org)
with your key address and we'll send you some `fermion` testnet tokens :)

Fetch and navigate to the repository of testnet data 
```
git clone https://github.com/tendermint/testnets $HOME/testnets
GAIANET=$HOME/testnets/atlas/gaia
cd $GAIANET
```

Now we can start a new node in the background, note that it may take a decent 
chunk of time to sync with the existing testnet... go brew a pot of coffee! 

```
gaia start --home=$GAIANET  &> atlas.log &
```

The above command will automaticaly generate a validator private key found in
`$GAIANET/priv_validator.json`. Let's get the pubkey data for our validator
node. The pubkey is located under `"pub_key"{"data":` within the json file

```
cat $GAIANET/priv_validator.json 
PUBKEY=<your newly generated address>  
```

If you have a json parser like `jq`, you can get just the pubkey:

```
PUBKEY=$(cat $GAIANET/priv_validator.json | jq -r .pub_key.data)
```

Next let's initialize the gaia client to start interacting with the testnet:

```
gaiacli init --chain-id=atlas --node=tcp://localhost:46657
```

First let's check out our balance

```
gaiacli query account $MYADDR
```

We are now ready to bond some tokens:

```
gaiacli tx bond --amount=5fermion --name=$MYNAME --pubkey=$PUBKEY
```

Bonding tokens means that your balance is tied up as _stake_. Don't worry,
you'll be able to get it back later. As soon as some tokens have been bonded
the validator node which we started earlier will have power in the network and
begin to participate in consensus!

We can now check the validator set and see that we are a part of the club!
```
gaiacli query validators
```

Finally lets unbond to get back our tokens

```
gaiacli tx unbond --amount=5fermion --name=$MYNAME
```

Remember to unbond before stopping your node!

### Local-Test Example

Here is a quick example to get you off your feet: 

First, generate a new key with a name, and save the address:

```
MYNAME=<your name>
gaiacli keys new $MYNAME
gaiacli keys list
MYADDR=<your newly generated address>
```
Now initialize a gaia chain:

```
gaia init $MYADDR --home=$HOME/.atlas1 --chain-id=test 
```

This will create all the files necessary to run a single node chain in `$HOME/.atlas1`:
a `priv_validator.json` file with the validators private key, and a `genesis.json` file 
with the list of validators and accounts. In this case, we have one random validator,
and `$MYADDR` is an independent account that has a bunch of coins.

We can add a second node on our local machine by initiating a node in a new directory,
and copying in the genesis:


```
gaia init $MYADDR --home=$HOME/.atlas2 --chain-id=test
cp $HOME/.atlas1/genesis.json $HOME/.atlas2/genesis.json
```

We need to also modify `$HOME/.atlas2/config.toml` to set new seeds and ports. It should look like:

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
seeds = "0.0.0.0:46656"
```

Great, now that we've initialized the chains, we can start both nodes in the background:

```
gaia start --home=$HOME/.atlas1  &> atlas1.log &
NODE1_PID=$!
gaia start --home=$HOME/.atlas2  &> atlas2.log &
NODE2_PID=$!
```

Note we save the `PID` so we can later kill the processes.

Of course, you can peak at your logs with `tail atlas1.log`, or follow them 
for a bit with `tail -f atlas1.log`.

Now we can initialize a client for the first node, and look up our account:

```
gaiacli init --chain-id=test --node=tcp://localhost:46657
gaiacli query account $MYADDR
```

Nice. We can also lookup the validator set:

```
gaiacli query validators
```

Notice it's empty! This is because the initial validators are special - 
the app doesn't know about them, so they can't be removed. To see what
tendermint itself thinks the validator set is, use:

```
curl localhost:46657/validators
```

Ok, let's add the second node as a validator. First, we need the pubkey data:

```
cat $HOME/.atlas2/priv_validator.json 
```

If you have a json parser like `jq`, you can get just the pubkey:

```
cat $HOME/.atlas2/priv_validator.json | jq .pub_key.data
```

Now we can bond some coins to that pubkey:

```
gaiacli tx bond --amount=10strings --name=$MYNAME --pubkey=<validator pubkey>
```

We should see our account balance decrement, and the pubkey get added to the app's list of bonds:

```
gaiacli query account $MYADDR
gaiacli query validators
``` 

To confirm for certain the new validator is active, check tendermint:

```
curl localhost:46657/validators
```

If you now kill your second node, blocks will stop streaming in, because there aren't enough validators online.
Turn her back on and they will start streaming again.

Finally, to relinquish all your power, unbond some coins. You should see your
VotingPower reduce and your account balance increase.

```
gaiacli tx unbond --amount=1strings --name=$MYNAME
gaiacli query validators
gaiacli query account $MYADDR
``` 

Once you unbond enough, you will no longer be needed to make new blocks.
