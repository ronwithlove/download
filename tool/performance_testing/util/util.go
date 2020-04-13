package util_ron

import (
	"errors"
	"fmt"
	"github.com/performance-testing-tool/go-dappley/common"
	"github.com/performance-testing-tool/go-dappley/core/account"
	"github.com/performance-testing-tool/go-dappley/core/transaction"
	"github.com/performance-testing-tool/go-dappley/core/utxo"
	utxopb "github.com/performance-testing-tool/go-dappley/core/utxo/pb"
	"github.com/performance-testing-tool/go-dappley/logic/ltransaction"
	rpcpb "github.com/performance-testing-tool/go-dappley/rpc/pb"
)


func CreateTransaction(respon *rpcpb.GetUTXOResponse, amount uint64,fromAccount,toAccount *account.Account)(transaction.Transaction, error){
	//从服务器返回的utxo集合里找到满足转账所需金额的utxo
	tx_utxos, err := getUTXOsWithAmount(
		respon.GetUtxos(),
		common.NewAmount(amount),
		common.NewAmount(0),
		common.NewAmount(0),
		common.NewAmount(0))
	if err != nil {
		fmt.Println("Error:", err.Error())
		return transaction.Transaction{}, err
	}
	//组装交易参数
	sendTxParam := transaction.NewSendTxParam(
		account.NewAddress(fromAccount.GetAddress().String()),
		fromAccount.GetKeyPair(),
		account.NewAddress(toAccount.GetAddress().String()),
		common.NewAmount(amount),
		common.NewAmount(0),
		common.NewAmount(0),
		common.NewAmount(0),
		"")

	return ltransaction.NewUTXOTransaction(tx_utxos, sendTxParam)

}

//从服务器返回的utxo集合里找到满足转账所需金额的utxo
func getUTXOsWithAmount(responUtxos []*utxopb.Utxo, amount *common.Amount, tip *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount) ([]*utxo.UTXO, error) {
	//得到Utxo集合
	var inputUtxos []*utxo.UTXO
	for _, u := range responUtxos{
		utxo := utxo.UTXO{}
		utxo.Value = common.NewAmountFromBytes(u.Amount)
		utxo.Txid = u.Txid
		utxo.PubKeyHash = account.PubKeyHash(u.PublicKeyHash)
		utxo.TxIndex = int(u.TxIndex)
		inputUtxos = append(inputUtxos, &utxo)
	}


	if tip != nil {
		amount = amount.Add(tip)
	}
	if gasLimit != nil {
		limitedFee := gasLimit.Mul(gasPrice)
		amount = amount.Add(limitedFee)
	}

	var retUtxos []*utxo.UTXO
	sum := common.NewAmount(0)
	for _, u := range inputUtxos {
		sum = sum.Add(u.Value)
		retUtxos = append(retUtxos, u)
		if sum.Cmp(amount) >= 0 {
			break
		}
	}

	if sum.Cmp(amount) < 0 {
		return nil, errors.New("cli: the balance is insufficient")
	}

	return retUtxos, nil
}

func GetAccountName(account *account.Account)(string){
	return account.GetAddress().String()[0:6]
}