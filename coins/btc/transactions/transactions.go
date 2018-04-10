package transactions

import (
	"sort"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/shiftdevices/godbb/coins/btc/blockchain"
	"github.com/shiftdevices/godbb/coins/btc/electrum/client"
	"github.com/shiftdevices/godbb/coins/btc/synchronizer"
	"github.com/shiftdevices/godbb/util/errp"
	"github.com/shiftdevices/godbb/util/locker"
	"github.com/sirupsen/logrus"
)

// TxOut is a transaction output which is part of the wallet.
type TxOut struct {
	*wire.TxOut
	Address btcutil.Address
}

// TxIn is a transaction input which is part of the wallet. The transaction hash of the transaction
// this input was found in is also recorded.
type TxIn struct {
	*wire.TxIn
	txHash chainhash.Hash
}

// Transaction is a transaction touching the wallet.
type Transaction struct {
	TX        *wire.MsgTx
	Height    int
	addresses map[string]struct{}
}

// Transactions handles wallet transactions: keeping an index of the transactions, inputs, (unspent)
// outputs, etc.
type Transactions struct {
	locker.Locker

	net            *chaincfg.Params
	transactions   map[chainhash.Hash]*Transaction
	requestedTXs   map[chainhash.Hash][]func(*wire.MsgTx)
	addressHistory map[string]map[chainhash.Hash]struct{}
	// inputs contains all inputs of all transactions that touch our wallet. It includes all inputs
	// that spend an output of our wallet, and more.  The inputs are referenced by the outputs they
	// spend.
	//
	// TODO: store slice of inputs along with the txhash they appear in. If there are more than one,
	// a double spend is detected.
	inputs map[wire.OutPoint]*TxIn
	// outputs contains all outputs which belong to the wallet.
	outputs map[wire.OutPoint]*TxOut

	synchronizer *synchronizer.Synchronizer
	blockchain   blockchain.Interface
	log          *logrus.Entry
}

// NewTransactions creates a new instance of Transactions.
func NewTransactions(
	net *chaincfg.Params,
	synchronizer *synchronizer.Synchronizer,
	blockchain blockchain.Interface,
	log *logrus.Entry,
) *Transactions {
	return &Transactions{
		net:            net,
		transactions:   map[chainhash.Hash]*Transaction{},
		addressHistory: map[string]map[chainhash.Hash]struct{}{},
		requestedTXs:   map[chainhash.Hash][]func(*wire.MsgTx){},
		outputs:        map[wire.OutPoint]*TxOut{},
		inputs:         map[wire.OutPoint]*TxIn{},

		synchronizer: synchronizer,
		blockchain:   blockchain,
		log:          log.WithFields(logrus.Fields{"group": "transactions", "net": net.Name}),
	}
}

func (transactions *Transactions) processTxForAddress(
	address btcutil.Address, txHash chainhash.Hash, tx *wire.MsgTx, height int) {
	// Don't process the tx if it is not found in the address history. It could have been removed
	// from the history before this function was called.
	if _, found := transactions.addressHistory[address.String()][txHash]; !found {
		return
	}

	transaction, ok := transactions.transactions[txHash]
	if !ok {
		transaction = &Transaction{
			TX:        tx,
			Height:    height,
			addresses: map[string]struct{}{},
		}
		transactions.transactions[txHash] = transaction
	}

	transaction.addresses[address.String()] = struct{}{}
	transaction.Height = height
	transactions.processInputsAndOutputsForAddress(address, txHash, tx)
}

// Go through the tx and extract all inputs and outputs which touch the address.
func (transactions *Transactions) processInputsAndOutputsForAddress(
	address btcutil.Address,
	txHash chainhash.Hash,
	tx *wire.MsgTx) {
	// Gather transaction inputs that spend outputs of the given address.
	for _, txIn := range tx.TxIn {
		// Since transactions can be processed in any order, and we might process the same tx
		// multiple times for different addresses, we index all inputs, even those that didn't
		// originate from our wallet. At this stage we don't know if it is one of our own inputs,
		// since the output that it spends might be indexed later.
		if existingTxIn, ok := transactions.inputs[txIn.PreviousOutPoint]; ok && existingTxIn.txHash != txHash {
			transactions.log.WithFields(logrus.Fields{"txIn.PreviousOutPoint": txIn.PreviousOutPoint,
				"existingTxIn.txHash": existingTxIn.txHash, "txHash": txHash}).
				Warning("Double spend detected")
		}
		transactions.inputs[txIn.PreviousOutPoint] = &TxIn{
			TxIn:   txIn,
			txHash: txHash,
		}
	}
	// Gather transaction outputs that belong to us.
	for index, txOut := range tx.TxOut {
		scriptClass, addresses, _, err := txscript.ExtractPkScriptAddrs(
			txOut.PkScript,
			transactions.net,
		)
		// We don't care about the script type, as the address alone uniquely identifies it. If the
		// address belongs to the wallet, it already knows what kind of output it is.
		_ = scriptClass
		if err != nil {
			transactions.log.WithField("error", err).Debug("Failed to extract pk script addresses")
			// Unrecognized output. Skip.
			continue
		}
		// For now we only look at single-address outputs (no multisig or other special contracts).
		if len(addresses) != 1 {
			transactions.log.WithField("addresses-length", len(addresses)).
				Debug("Only supporting single-address outputs for now")
			continue
		}
		// Check if output is ours.
		if addresses[0].String() == address.String() {
			transactions.outputs[wire.OutPoint{
				Hash:  txHash,
				Index: uint32(index),
			}] = &TxOut{
				TxOut:   txOut,
				Address: address,
			}
		}
	}
}

func (transactions *Transactions) allInputsOurs(transaction *wire.MsgTx) bool {
	for _, txIn := range transaction.TxIn {
		if _, ok := transactions.outputs[txIn.PreviousOutPoint]; !ok {
			return false
		}
	}
	return true
}

// SpendableOutputs returns all unspent outputs of the wallet which are eligible to be spent. Those
// include all unspent outputs of confirmed transactions, and unconfirmed outputs that we created
// ourselves.
func (transactions *Transactions) SpendableOutputs() map[wire.OutPoint]*TxOut {
	transactions.synchronizer.WaitSynchronized()
	defer transactions.RLock()()
	result := map[wire.OutPoint]*TxOut{}
	for outPoint, txOut := range transactions.outputs {
		tx := transactions.transactions[outPoint.Hash]
		confirmed := tx.Height > 0
		_, spent := transactions.inputs[outPoint]
		if !spent && (confirmed || transactions.allInputsOurs(tx.TX)) {
			result[outPoint] = txOut
		}
	}
	return result
}

func (transactions *Transactions) removeTxForAddress(address btcutil.Address, txHash chainhash.Hash) {
	transactions.log.Debug("Remove transaction for address")
	transaction, ok := transactions.transactions[txHash]
	if !ok {
		// Not yet indexed.
		transactions.log.Debug("Transaction hash not listed")
		return
	}

	delete(transaction.addresses, address.String())
	transactions.log.Debug("Deleting transaction address")
	if len(transaction.addresses) == 0 {
		// Tx is not touching any of our outputs anymore. Remove.

		for _, txIn := range transaction.TX.TxIn {
			delete(transactions.inputs, txIn.PreviousOutPoint)
			transactions.log.Debug("Deleting transaction output")
		}

		// Remove the outputs added by this tx.
		for index := range transaction.TX.TxOut {
			outPoint := wire.OutPoint{
				Hash:  txHash,
				Index: uint32(index),
			}
			if _, ok := transactions.outputs[outPoint]; ok {
				delete(transactions.outputs, outPoint)
			}
		}

		delete(transactions.transactions, txHash)
	}
}

// UpdateAddressHistory should be called when initializing a wallet address, or when the history of
// an address changes (a new transaction that touches it appears or disappears). The transactions
// are downloaded and indexed.
func (transactions *Transactions) UpdateAddressHistory(address btcutil.Address, txs []*client.TxInfo) {
	defer transactions.Lock()()
	txsSet := map[chainhash.Hash]struct{}{}
	for _, txInfo := range txs {
		txsSet[txInfo.TXHash.Hash()] = struct{}{}
	}
	if len(txsSet) != len(txs) {
		err := errp.New("duplicate tx ids in address history returned by server")
		transactions.log.WithField("error", err).Panic(err)
		// TODO
		panic(err)
	}
	for txHash := range transactions.addressHistory[address.String()] {
		if _, txOK := txsSet[txHash]; txOK {
			continue
		}
		// A tx was previously in the address history but is not anymore.  If the tx was already
		// downloaded and indexed, it will be removed.  If it is currently downloading (enqueued for
		// indexing), it will not be processed.
		transactions.removeTxForAddress(address, txHash)
	}

	transactions.addressHistory[address.String()] = txsSet

	for _, txInfo := range txs {
		func(txHash chainhash.Hash, height int) {
			transactions.doForTransaction(txHash, func(tx *wire.MsgTx) {
				transactions.processTxForAddress(address, txHash, tx, height)
			})
		}(txInfo.TXHash.Hash(), txInfo.Height)
	}
}

func (transactions *Transactions) doForTransaction(
	txHash chainhash.Hash,
	callback func(tx *wire.MsgTx),
) {
	tx, ok := transactions.transactions[txHash]
	if ok {
		callback(tx.TX)
		return
	}
	if transactions.requestedTXs[txHash] == nil {
		transactions.requestedTXs[txHash] = []func(*wire.MsgTx){}
	}
	alreadyDownloading := len(transactions.requestedTXs[txHash]) != 0
	transactions.requestedTXs[txHash] = append(transactions.requestedTXs[txHash], callback)
	if alreadyDownloading {
		return
	}
	done := transactions.synchronizer.IncRequestsCounter()
	if err := transactions.blockchain.TransactionGet(
		txHash,
		func(tx *wire.MsgTx) error {
			defer transactions.Lock()()
			for _, callback := range transactions.requestedTXs[txHash] {
				callback(tx)
			}
			delete(transactions.requestedTXs, txHash)
			return nil
		},
		done,
	); err != nil {
		transactions.log.WithField("error", err).Panic("Failed to retrieve transaction")
		// TODO
		panic(err)
	}
}

// Balance contains the available and incoming balance of the wallet.
type Balance struct {
	// Available funds are all confirmed funds which are not spent by any tx. Exception: unconfirmed
	// transactions that spend from the wallet are available.
	Available btcutil.Amount
	// Incoming balance are unconfirmed funds coming into the wallet.
	Incoming btcutil.Amount
}

// Balance computes the confirmed and unconfirmed balance of the wallet.
func (transactions *Transactions) Balance() *Balance {
	transactions.synchronizer.WaitSynchronized()
	defer transactions.RLock()()
	var available, incoming int64
	for outPoint, txOut := range transactions.outputs {
		// What is spent can not be available nor incoming.
		if _, spent := transactions.inputs[outPoint]; spent {
			continue
		}
		tx := transactions.transactions[outPoint.Hash]
		confirmed := tx.Height > 0
		if confirmed || transactions.allInputsOurs(tx.TX) {
			available += txOut.Value
		} else {
			incoming += txOut.Value
		}
	}
	return &Balance{
		Available: btcutil.Amount(available),
		Incoming:  btcutil.Amount(incoming),
	}
}

// byHeight defines the methods needed to satisify sort.Interface to sort transactions by their
// height. Special case for unconfirmed transactions (height <=0), which come last.
type byHeight []*TxInfo

func (s byHeight) Len() int { return len(s) }
func (s byHeight) Less(i, j int) bool {
	if s[j].Height <= 0 {
		return true
	}
	if s[i].Height <= 0 {
		return false
	}
	return s[i].Height < s[j].Height
}
func (s byHeight) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// TxType is a type of transaction. See the TxType* constants.
type TxType string

const (
	// TxTypeReceive is a tx which sends funds to our wallet.
	TxTypeReceive TxType = "receive"
	// TxTypeSend is a tx which sends funds out of our wallet.
	TxTypeSend = "send"
	// TxTypeSendSelf is a tx from out wallet to our wallet.
	TxTypeSendSelf = "sendSelf"
)

// TxInfo contains additional tx information to display to the user.
type TxInfo struct {
	*Transaction
	Type TxType
	// Amount is always >0 and is the amount received or sent (not including the fee).
	Amount btcutil.Amount
	// Fee is nil if for a receiving tx (TxTypeReceive). The fee is only displayed (and relevant)
	// when sending funds from the wallet.
	Fee *btcutil.Amount
}

// txInfo computes additional information to display to the user (type of tx, fee paid, etc.).
func (transactions *Transactions) txInfo(
	tx *Transaction,
	isChangeAddress func(btcutil.Address) bool) *TxInfo {
	defer transactions.RLock()()
	var sumOurInputs btcutil.Amount
	var result btcutil.Amount
	allInputsOurs := true
	for _, txIn := range tx.TX.TxIn {
		if spentOut, ok := transactions.outputs[txIn.PreviousOutPoint]; ok {
			sumOurInputs += btcutil.Amount(spentOut.Value)
		} else {
			allInputsOurs = false
		}
	}
	var sumAllOutputs, sumOurReceive, sumOurChange btcutil.Amount
	allOutputsOurs := true
	for index, txOut := range tx.TX.TxOut {
		sumAllOutputs += btcutil.Amount(txOut.Value)
		if output, ok := transactions.outputs[wire.OutPoint{
			Hash:  tx.TX.TxHash(),
			Index: uint32(index),
		}]; ok {
			if isChangeAddress(output.Address) {
				sumOurChange += btcutil.Amount(txOut.Value)
			} else {
				sumOurReceive += btcutil.Amount(txOut.Value)
			}
		} else {
			allOutputsOurs = false
		}
	}
	var txType TxType
	var feeP *btcutil.Amount
	if allInputsOurs {
		fee := sumOurInputs - sumAllOutputs
		feeP = &fee
		if allOutputsOurs {
			txType = TxTypeSendSelf
			// Money sent from our wallet to our wallet
			result = sumOurReceive
		} else {
			// Money sent from our wallet to external address.
			txType = TxTypeSend
			result = sumAllOutputs - sumOurReceive - sumOurChange
		}
	} else {
		// Money sent from external to our wallet
		txType = TxTypeReceive
		result = sumOurReceive + sumOurChange - sumOurInputs
	}
	return &TxInfo{
		Transaction: tx,
		Type:        txType,
		Amount:      result,
		Fee:         feeP,
	}
}

// Transactions returns an ordered list of transactions.
func (transactions *Transactions) Transactions(
	isChangeAddress func(btcutil.Address) bool) []*TxInfo {
	transactions.synchronizer.WaitSynchronized()
	defer transactions.RLock()()
	txs := []*TxInfo{}
	for _, transaction := range transactions.transactions {
		txs = append(txs, transactions.txInfo(transaction, isChangeAddress))
	}
	sort.Sort(sort.Reverse(byHeight(txs)))
	return txs
}
