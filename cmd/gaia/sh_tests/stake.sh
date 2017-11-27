#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE="gaia node"
CLIENT_EXE="gaia client"
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}

oneTimeSetUp() {
    #[ "$2" ] || echo "missing parameters, line=${LINENO}" ; exit 1;

    # These are passed in as args
    BASE_DIR=$HOME/stake_test
    CHAIN_ID="stake_test"

    # TODO Make this more robust
    if [ "$BASE_DIR" == "$HOME/" ]; then
        echo "Must be called with argument, or it will wipe your home directory"
        exit 1
    fi

    rm -rf $BASE_DIR 2>/dev/null
    mkdir -p $BASE_DIR

    # Set up client - make sure you use the proper prefix if you set
    #   a custom CLIENT_EXE
    export BC_HOME=${BASE_DIR}/client
    prepareClient

    # start basecoin server (with counter)
    initServer $BASE_DIR $CHAIN_ID
    if [ $? != 0 ]; then return 1; fi

    initClient $CHAIN_ID
    if [ $? != 0 ]; then return 1; fi

    printf "...Testing may begin!\n\n\n"

}

oneTimeTearDown() {
    quickTearDown
}

test00GetAccount() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    assertFalse "line=${LINENO}, requires arg" "${CLIENT_EXE} query account"

    checkAccount $SENDER "9007199254740992"

    ACCT2=$(${CLIENT_EXE} query account $RECV 2>/dev/null)
    assertFalse "line=${LINENO}, has no genesis account" $?
}

test01SendTx() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    assertFalse "line=${LINENO}, missing dest" "${CLIENT_EXE} tx send --amount=992fermion --sequence=1"
    assertFalse "line=${LINENO}, bad password" "echo foo | ${CLIENT_EXE} tx send --amount=992fermion --sequence=1 --to=$RECV --name=$RICH"
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992fermion --sequence=1 --to=$RECV --name=$RICH)
    txSucceeded $? "$TX" "$RECV"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    checkAccount $SENDER "9007199254740000" $TX_HEIGHT
    # make sure 0x prefix also works
    checkAccount "0x$SENDER" "9007199254740000" $TX_HEIGHT
    checkAccount $RECV "992" $TX_HEIGHT

    # Make sure tx is indexed
    checkSendTx $HASH $TX_HEIGHT $SENDER "992"
}

# Load common then run these tests with shunit2!
CLI_DIR=$GOPATH/src/github.com/cosmos/gaia/vendor/github.com/cosmos/cosmos-sdk/tests/cli

. $CLI_DIR/common.sh
. $CLI_DIR/shunit2
