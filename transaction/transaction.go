package transaction

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/uptrace/bun"
	"golang.org/x/sync/semaphore"
	"math"
	"math/big"
	"time"
)

type TrackingTransaction struct {
	Address common.Address
}

type Transaction struct {
	From      string
	To        string
	RawAmount string
	Amount    float64
	Hash      string
	Status    string
	Time      time.Time
	Block     int64
	Event     []Event
}

type Event struct {
	From         string
	To           string
	Token        string
	Amount       float64
	RawAmount    string
	Hash         string
	Status       string
	TokenAddress string
}

func NewTxTracking(address string) (*TrackingTransaction, error) {
	addr := common.HexToAddress(address)
	tracking := &TrackingTransaction{
		Address: addr,
	}

	return tracking, nil
}

func (tracking *TrackingTransaction) GetTransactionFromBlockNumber(db *bun.DB, client *ethclient.Client, number *big.Int, msg chan string) {
	sem := semaphore.NewWeighted(2)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	for i := number.Int64(); ; i++ {
		blockNumber := big.NewInt(i)

		err := sem.Acquire(context.Background(), 1)
		if err != nil {
			fmt.Println(err)
		}

		go func() {
			data, err := tracking.GetTransactionByBlockNumber(client, chainID, blockNumber)
			if err != nil {
				fmt.Println(blockNumber, err)
			}

			err = tracking.InsertTransactionToDb(db, data)
			if err != nil {
				fmt.Println(blockNumber, err)
			}

			tracking.SendTransactionToBot(msg, data)
			sem.Release(1)
		}()
	}
}

func (tracking *TrackingTransaction) GetTransactionByBlockNumber(client *ethclient.Client, chainID *big.Int, number *big.Int) ([]Transaction, error) {
	lastestBlockNumber, err := GetLatestBlockNumber(client)
	if err != nil {
		return nil, err
	}

	if number.Int64() > lastestBlockNumber.Int64() {
		timeDelay := (number.Int64()-lastestBlockNumber.Int64())*3 + 1
		time.Sleep(time.Duration(timeDelay) * time.Second)
	}

	block, err := client.BlockByNumber(context.Background(), number)
	if err != nil {
		return nil, err
	}

	var data []Transaction
	for _, tx := range block.Transactions() {
		msg, err := tx.AsMessage(types.NewEIP155Signer(chainID), tx.GasPrice())
		if err != nil {
			return nil, err
		}

		isInTransaction, status := tracking.IsInTransaction(msg, tx)
		if isInTransaction == false {
			continue
		}

		event, err := tracking.GetEventInTx(client, tx.Hash())
		if err != nil {
			return nil, err
		}

		recipient := GetRecipient(tx)
		timestamp := time.Unix(int64(block.Time()), 0)
		data = append(data, Transaction{
			From:      msg.From().Hex(),
			To:        recipient,
			RawAmount: tx.Value().String(),
			Amount:    AmountFloatPrecision(tx.Value(), 6),
			Hash:      tx.Hash().Hex(),
			Status:    *status,
			Time:      timestamp,
			Block:     block.Number().Int64(),
			Event:     *event,
		})
	}
	return data, nil
}

var logTransferSigHash = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

func (tracking *TrackingTransaction) GetEventInTx(client *ethclient.Client, hash common.Hash) (*[]Event, error) {
	receipt, err := client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, err
	}

	event := new([]Event)

	for _, vLog := range receipt.Logs {
		switch vLog.Topics[0].Hex() {
		case logTransferSigHash.Hex():
			err = tracking.GetTransferEventData(client, *vLog, event, hash)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	return event, nil
}

func (tracking *TrackingTransaction) GetTransferEventData(client *ethclient.Client, vLog types.Log, event []*Event, hash common.Hash) error {
	data := new(big.Int)
	fromAddress := common.HexToAddress(vLog.Topics[1].Hex())
	toAddress := common.HexToAddress(vLog.Topics[2].Hex())

	statusOfTransaction := "in"
	if tracking.Address.String() == fromAddress.String() {
		statusOfTransaction = "out"
	}
	isInEvent, _ := tracking.IsInEvent(fromAddress, toAddress)
	if isInEvent == false {
		return nil
	}

	token, err := GetTokenName(client, vLog.Address)
	if err != nil {
		return err
	}

	amount := data.SetBytes(vLog.Data)

	event = append(event, &Event{
		From:         fromAddress.String(),
		To:           toAddress.String(),
		Token:        token,
		RawAmount:    amount.String(),
		Amount:       AmountFloatPrecision(amount, 6),
		Hash:         hash.String(),
		Status:       statusOfTransaction,
		TokenAddress: vLog.Address.String(),
	})

	return nil
}

func (tracking *TrackingTransaction) IsInEvent(fromAddress common.Address, toAddress common.Address) (bool, *string) {
	if fromAddress != tracking.Address && toAddress != tracking.Address {
		return false, nil
	}

	var status string
	if fromAddress == tracking.Address {
		status = "out"
		return true, &status
	}

	status = "in"
	return true, &status
}

func GetLatestBlockNumber(client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return header.Number, nil
}

func (tracking *TrackingTransaction) IsInTransaction(msg types.Message, tx *types.Transaction) (bool, *string) {
	var status string
	if msg.From().Hex() == tracking.Address.String() {
		status = "out"
		return true, &status
	}

	if tx.To() == nil {
		return false, nil
	}

	if tx.To().Hex() == tracking.Address.String() {
		status = "in"
		return true, &status
	}
	return false, nil
}

func GetRecipient(tx *types.Transaction) string {
	if tx.To() == nil {
		return "null"
	}

	return tx.To().String()
}

func GetTokenName(client *ethclient.Client, address common.Address) (string, error) {
	instance, err := NewToken(address, client)
	if err != nil {
		return "", err
	}

	name, err := instance.Symbol(&bind.CallOpts{})
	if err != nil {
		return "", err
	}

	return name, nil
}

func AmountFloatPrecision(num *big.Int, precision int) float64 {
	// pow18 = 10 ** 18 in BigInt type
	divisor := new(big.Int)
	divisor.Exp(big.NewInt(10), big.NewInt(18), nil)

	n := new(big.Float).SetInt(num)
	n.Quo(n, new(big.Float).SetInt(divisor))
	// p = 10 ** precision in float64 type
	p := math.Pow10(precision) // big float

	tmp, _ := n.Float64()
	value := float64(int(tmp*p)) / p
	return value
}
