package stake

import (
	"fmt"
	"strconv"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

//nolint
const (
	name = "stake"

	queueUnbondTypeByte = iota
	queueCommissionTypeByte
)

//nolint
var (
	//TODO should all these global parameters be moved to the state?
	periodUnbonding uint64 = 30               // queue blocks before unbond
	bondDenom       string = "atom"           // bondable coin denomination
	maxVal          int    = 100              // maximum number of validators
	minValBond             = NewDecimal(5, 5) // minumum number of bonded coins to be a validator

	maxCommHistory           = NewDecimal(5, -2) // maximum total commission permitted across the queued commission history
	periodCommHistory uint64 = 28800             // 1 day @ 1 block/3 sec

	inflationPerReward Decimal = NewDecimal(6654498, -15) // inflation between (0 to 1). ~7% annual at 1 block/3sec

	// GasAllocations per staking transaction
	costBond     = uint64(20)
	costUnbond   = uint64(0)
	costNominate = uint64(20)
	costModComm  = uint64(20)
)

// Name - simply the name TODO do we need name to be unexposed for security?
func Name() string {
	return name
}

// Handler - the transaction processing handler
type Handler struct {
	stack.PassInitValidate
}

var _ stack.Dispatchable = Handler{} // enforce interface at compile time

// Name - return stake namespace
func (Handler) Name() string {
	return name
}

// AssertDispatcher - placeholder for stack.Dispatchable
func (Handler) AssertDispatcher() {}

// InitState - set genesis parameters for staking
func (Handler) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, cb sdk.InitStater) (log string, err error) {
	return "", initState(module, key, value, store)
}

//separated for testing
func initState(module, key, value string, store state.SimpleDB) (err error) {
	if module != name {
		return errors.ErrUnknownModule(module)
	}
	switch key {
	case "bond_coin":
		bondDenom = value
	case "maxval",
		"commhist_period",
		"unbond_period",
		"atomsupply",
		"minvalbond",
		"cost_bond",
		"cost_unbond",
		"cost_nominate",
		"cost_modcomm":
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("input must be integer, Error: %v", err.Error())
		}
		switch key {
		case "maxval":
			maxVal = i
		case "commhist_period":
			periodCommHistory = uint64(i)
		case "unbond_period":
			periodUnbonding = uint64(i)
		case "atomsupply":
			saveAtomSupply(store, NewDecimal(int64(i), 0))
		case "minvalbond":
			minValBond = NewDecimal(int64(i), 0)
		case "cost_bond":
			costBond = uint64(i)
		case "cost_unbond":
			costUnbond = uint64(i)
		case "cost_nominate":
			costNominate = uint64(i)
		case "cost_modcomm":
			costModComm = uint64(i)
		}
	}
	return errors.ErrUnknownKey(key)
}

// CheckTx checks if the tx is properly structured
func (h Handler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, _ sdk.Checker) (res sdk.CheckResult, err error) {

	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	switch t := tx.Unwrap().(type) {
	case TxBond:
		// TODO could also check for existence of delegatee here? (already checked in deliverTx)

		//check for suffient funds to send
		sender, abciRes := getSingleSender(ctx)
		if abciRes.IsErr() {
			return res, abciRes
		}
		_, err := coin.CheckCoins(store, sender, coin.Coins{t.Amount}.Negative())
		if err != nil {
			return res, err
		}
		return sdk.NewCheck(costBond, ""), nil

	case TxUnbond:
		// TODO check for sufficient delegatee tokens to unbond here? (already checked in deliverTx)
		return sdk.NewCheck(costUnbond, ""), nil

	case TxNominate:
		// TODO check if already a delegatee? (already checked in deliverTx)
		if !ctx.HasPermission(t.Nominee) {
			return res, errors.ErrUnauthorized()
		}
		return sdk.NewCheck(costNominate, ""), nil

	case TxModComm:
		// TODO check if is a delegatee? (already checked in deliverTx)
		if !ctx.HasPermission(t.Delegatee) {
			return res, errors.ErrUnauthorized()
		}
		return sdk.NewCheck(costModComm, ""), nil
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// Tick - Called every block even if no transaction,
//   process all queues, validator rewards, and calculate the validator set difference
func (h Handler) Tick(ctx sdk.Context, height uint64, store state.SimpleDB,
	dispatch sdk.Deliver) (diffVal []*abci.Validator, err error) {

	// Process the unbonding queue
	sendCoins := func(sender, receiver sdk.Actor, amount coin.Coins) (err error) {
		tx := coin.NewSendOneTx(sender, receiver, amount)

		//TODO Context should be created within sendcoins based on the send??
		_, err = dispatch.DeliverTx(ctx, store, tx)
		return
	}
	err = processQueueUnbond(sendCoins, store, height)
	if err != nil {
		return
	}

	//process the historical commission changes queue
	err = processQueueCommHistory(store, height)
	if err != nil {
		return
	}

	// Determine the validator set changes
	delegateeBonds, err := loadDelegateeBonds(store)
	if err != nil {
		return
	}
	startVal := delegateeBonds.GetValidators()
	totalVotingPower := delegateeBonds.UpdateVotingPower()
	newVal := delegateeBonds.GetValidators()
	diffVal = ValidatorsDiff(startVal, newVal)

	// Process validator set rewards
	creditCoins := func(receiver sdk.Actor, amount coin.Coins) (err error) {
		tx := coin.NewCreditTx(receiver, amount)
		_, err = dispatch.DeliverTx(ctx, store, tx)
		return
	}
	err = processValidatorRewards(creditCoins, store, height, totalVotingPower)
	if err != nil {
		return
	}

	return
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, dispatch sdk.Deliver) (res sdk.DeliverResult, err error) {

	_, err = h.CheckTx(ctx, store, tx, nil)
	if err != nil {
		return
	}

	//Run the transaction
	unwrap := tx.Unwrap()
	var abciRes abci.Result
	switch txType := unwrap.(type) {
	case TxBond:
		abciRes = runTxBond(ctx, store, txType, dispatch)
	case TxUnbond:
		abciRes = runTxUnbond(ctx, store, txType)
	case TxNominate:
		abciRes = runTxNominate(ctx, store, txType, dispatch)
	case TxModComm:
		abciRes = runTxModComm(ctx, store, txType)
	}

	res = sdk.DeliverResult{
		Data: abciRes.Data,
		Log:  abciRes.Log,
	}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func runTxBond(ctx sdk.Context, store state.SimpleDB, tx TxBond,
	dispatch sdk.Deliver) (res abci.Result) {

	sender, res := getSingleSender(ctx)
	if res.IsErr() {
		return res
	}

	sendCoins := func(receiver sdk.Actor, amount coin.Coins) (res abci.Result) {

		// Move coins from the deletator account to the delegatee lock account
		send := coin.NewSendOneTx(sender, receiver, amount)

		// If the deduction fails (too high), abort the command
		_, err := dispatch.DeliverTx(ctx, store, send)
		if err != nil {
			return abci.ErrInsufficientFunds.AppendLog(err.Error())
		}
		return
	}
	return runTxBondGuts(sendCoins, store, tx, sender)
}

func getSingleSender(ctx sdk.Context) (sender sdk.Actor, res abci.Result) {
	senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
	if len(senders) != 1 {
		return sender, resMissingSignature
	}
	return senders[0], abci.OK
}

func runTxBondGuts(sendCoins func(receiver sdk.Actor, amount coin.Coins) abci.Result,
	store state.SimpleDB, tx TxBond, sender sdk.Actor) abci.Result {

	// Get amount of coins to bond
	bondCoin := tx.Amount
	bondAmt := NewDecimal(bondCoin.Amount, 0)

	// Get the delegatee bond account
	delegateeBonds, err := loadDelegateeBonds(store)
	if err != nil {
		return resErrLoadingDelegatees(err)
	}
	i, delegateeBond := delegateeBonds.Get(tx.Delegatee)
	if delegateeBond == nil {
		return resBondNotNominated
	}

	// Move coins from the delegator account to the delegatee lock account
	res := sendCoins(delegateeBond.Account, coin.Coins{bondCoin})
	if res.IsErr() {
		return res
	}

	// Get or create delegator bonds
	delegatorBonds, err := loadDelegatorBonds(store, sender)
	if err != nil {
		return resErrLoadingDelegators(err)
	}
	if len(delegatorBonds) == 0 {
		delegatorBonds = DelegatorBonds{
			&DelegatorBond{
				Delegatee:  tx.Delegatee,
				BondTokens: Zero,
			},
		}
	}

	// Calculate amount of bond tokens to create, based on exchange rate
	bondTokens := bondAmt.Div(delegateeBond.ExchangeRate)
	j, _ := delegatorBonds.Get(tx.Delegatee)
	delegatorBonds[j].BondTokens = delegatorBonds[j].BondTokens.Add(bondTokens)
	delegateeBonds[i].TotalBondTokens = delegateeBonds[i].TotalBondTokens.Add(bondTokens)

	// NOTE the exchange rate does not change due to the bonding process,
	// all tokens are assigned AT the exchange rate
	// (but does change when validator rewards are processed)

	// Save to store
	saveDelegateeBonds(store, delegateeBonds)
	saveDelegatorBonds(store, sender, delegatorBonds)

	return abci.OK
}

func runTxUnbond(ctx sdk.Context, store state.SimpleDB, tx TxUnbond) (res abci.Result) {

	getSender := func() (sender sdk.Actor, res abci.Result) {
		senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
		if len(senders) != 0 {
			res = resMissingSignature
			return
		}
		sender = senders[0]
		return
	}

	return runTxUnbondGuts(getSender, store, tx, ctx.BlockHeight())
}

func runTxUnbondGuts(getSender func() (sdk.Actor, abci.Result), store state.SimpleDB, tx TxUnbond,
	height uint64) (res abci.Result) {

	bondAmt := NewDecimal(tx.Amount.Amount, 0)

	sender, res := getSender()
	if res.IsErr() {
		return res
	}

	//get delegator bond
	delegatorBonds, err := loadDelegatorBonds(store, sender)
	if err != nil {
		return resErrLoadingDelegators(err)
	}
	_, delegatorBond := delegatorBonds.Get(tx.Delegatee)
	if delegatorBond == nil {
		return resNoDelegatorForAddress
	}

	//get delegatee bond
	delegateeBonds, err := loadDelegateeBonds(store)
	if err != nil {
		return resErrLoadingDelegatees(err)
	}
	bvIndex, delegateeBond := delegateeBonds.Get(tx.Delegatee)
	if delegateeBond == nil {
		return resNoDelegateeForAddress
	}

	// subtract bond tokens from delegatorBond
	if delegatorBond.BondTokens.LT(bondAmt) {
		return resInsufficientFunds
	}
	delegatorBond.BondTokens = delegatorBond.BondTokens.Sub(bondAmt)

	if delegatorBond.BondTokens.Equal(Zero) {
		//begin to unbond all of the tokens if the validator unbonds their last token
		if sender.Equals(tx.Delegatee) {
			res = fullyUnbondDelegatee(delegateeBond, store, height)
			if res.IsErr() {
				return res //TODO add more context to this error?
			}
		} else {
			removeDelegatorBonds(store, sender)
		}
	} else {
		saveDelegatorBonds(store, sender, delegatorBonds)
	}

	// subtract tokens from delegateeBonds
	delegateeBond.TotalBondTokens = delegateeBond.TotalBondTokens.Sub(bondAmt)
	if delegateeBond.TotalBondTokens.Equal(Zero) {
		delegateeBonds.Remove(bvIndex)
	}

	// NOTE the exchange rate does not change due to the unbonding process
	// all tokens are unbonded AT the exchange rate
	// (but does change when validator rewards are processed)

	saveDelegateeBonds(store, delegateeBonds)

	// add unbond record to queue
	queueElem := QueueElemUnbond{
		QueueElem: QueueElem{
			Delegatee:    tx.Delegatee,
			HeightAtInit: height, // will unbond at `height + periodUnbonding`
		},
		Account:    sender,
		BondTokens: bondAmt,
	}
	queue, err := LoadQueue(queueUnbondTypeByte, store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error()) //should never occur
	}
	bytes := wire.BinaryBytes(queueElem)
	queue.Push(bytes)

	return abci.OK
}

//TODO improve efficiency of this function
func fullyUnbondDelegatee(delegateeBond *DelegateeBond, store state.SimpleDB, height uint64) (res abci.Result) {

	//TODO upgrade list queue... make sure that endByte as nil is okay
	allDelegators := store.List([]byte{delegatorKeyPrefix}, nil, maxVal)

	for _, delegatorRec := range allDelegators {

		delegator, err := getDelegatorFromKey(delegatorRec.Key)
		if err != nil {
			return resErrLoadingDelegator(delegatorRec.Key) //should never occur
		}
		delegatorBonds, err := loadDelegatorBonds(store, delegator)
		if err != nil {
			return resErrLoadingDelegators(err)
		}
		for _, delegatorBond := range delegatorBonds {
			if delegatorBond.Delegatee.Equals(delegateeBond.Delegatee) {
				getSender := func() (sdk.Actor, abci.Result) {
					return delegator, abci.OK
				}
				coinAmount := delegatorBond.BondTokens.Mul(delegateeBond.ExchangeRate)
				tx := NewTxUnbond(delegateeBond.Delegatee, coin.Coin{bondDenom,
					coinAmount.IntPart()}) //TODO conv to decimal
				res = runTxUnbondGuts(getSender, store, tx.Unwrap().(TxUnbond), height)
				if res.IsErr() {
					return res
				}
			}
		}
	}
	return abci.OK
}

func runTxNominate(ctx sdk.Context, store state.SimpleDB, tx TxNominate,
	dispatch sdk.Deliver) (res abci.Result) {

	bondCoins := func(bondAccount sdk.Actor, amount coin.Coins) abci.Result {
		senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
		if len(senders) == 0 {
			return resMissingSignature
		}
		send := coin.NewSendOneTx(senders[0], bondAccount, amount)
		_, err := dispatch.DeliverTx(ctx, store, send)
		if err != nil {
			return resInsufficientFunds
		}
		return abci.OK
	}
	return runTxNominateGuts(bondCoins, store, tx)
}

func runTxNominateGuts(bondCoins func(bondAccount sdk.Actor, amount coin.Coins) abci.Result,
	store state.SimpleDB, tx TxNominate) (res abci.Result) {

	// Create bond value object
	delegateeBond := DelegateeBond{
		Delegatee:       tx.Nominee,
		Commission:      tx.Commission,
		ExchangeRate:    One,
		TotalBondTokens: NewDecimal(int64(tx.Amount.Amount), 0), //TODO make coins decimal
		Account:         sdk.NewActor(name, append([]byte{0x00}, tx.Nominee.Address...)),
		VotingPower:     Zero,
	}

	delegatorBonds := DelegatorBonds{{
		Delegatee:  tx.Nominee,
		BondTokens: NewDecimal(int64(tx.Amount.Amount), 0), //TODO make coins decimal
	}}

	// Bond the coins to the bond account
	res = bondCoins(delegateeBond.Account, coin.Coins{tx.Amount})
	if res.IsErr() {
		return res
	}

	// Append and store to DelegateeBonds
	delegateeBonds, err := loadDelegateeBonds(store)
	if err != nil {
		return resErrLoadingDelegatees(err)
	}
	delegateeBonds = append(delegateeBonds, &delegateeBond)
	saveDelegateeBonds(store, delegateeBonds)

	// Also save a delegator account where the nominator
	// has delegated coins to their own delegatee account
	saveDelegatorBonds(store, tx.Nominee, delegatorBonds)

	return abci.OK
}

func runTxModComm(ctx sdk.Context, store state.SimpleDB, tx TxModComm) (res abci.Result) {
	return runTxModCommGuts(store, tx, ctx.BlockHeight())
}

func runTxModCommGuts(store state.SimpleDB, tx TxModComm, height uint64) (res abci.Result) {

	// Retrieve the record to modify
	delegateeBonds, err := loadDelegateeBonds(store)
	if err != nil {
		return resErrLoadingDelegatees(err)
	}
	record, delegateeBond := delegateeBonds.Get(tx.Delegatee)
	if delegateeBond == nil {
		return resNoDelegateeForAddress
	}

	// Determine that the amount of change proposed is permissible according to the queue change amount
	queue, err := LoadQueue(queueCommissionTypeByte, store)
	if err != nil {
		return resErrLoadingQueue(err) //should never occur
	}

	// First determine the sum of the changes in the queue
	// NOTE optimization opportunity: currently need to iterate through all records to pick out
	// the relevant records for the delegatee modifying their commission... Alternative would be to
	// have an array of all the commission change queues and process them individually in Tick
	// which would allow us to only grab one queue for here

	sumModCommQueue := Zero
	valuesBytes := queue.GetAll()
	for _, modCommBytes := range valuesBytes {

		var modComm QueueElemCommChange
		err = wire.ReadBinaryBytes(modCommBytes, modComm)
		if err != nil {
			return abci.ErrBaseEncodingError.AppendLog(err.Error()) //should never occur under normal operation
		}
		if modComm.Delegatee.Equals(tx.Delegatee) {
			sumModCommQueue.Add(modComm.CommChange)
		}
	}

	commChange := tx.Commission.Sub(delegateeBond.Commission)
	if sumModCommQueue.Add(commChange).GT(maxCommHistory) {
		return abci.ErrUnauthorized.AppendLog(
			fmt.Sprintf("proposed change is greater than is permissible. \n"+
				"The maximum change is %v per %v blocks, the your proposed change will "+
				"bring your queued change to %v", maxCommHistory, periodCommHistory, commChange))
	}

	// Execute the change and add the change amount to the queue
	// Retrieve, Modify and save the commission
	delegateeBonds[record].Commission = tx.Commission
	saveDelegateeBonds(store, delegateeBonds)

	// Add the commission modification the queue
	queueElem := QueueElemCommChange{
		QueueElem: QueueElem{
			Delegatee:    tx.Delegatee,
			HeightAtInit: height,
		},
		CommChange: tx.Commission,
	}
	bytes := wire.BinaryBytes(queueElem)
	queue.Push(bytes)

	return abci.OK
}
