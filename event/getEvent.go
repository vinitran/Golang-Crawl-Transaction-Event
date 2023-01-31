//package event
//
//import (
//	"context"
//	"ether/tables"
//	"fmt"
//	"github.com/ethereum/go-ethereum"
//	"github.com/ethereum/go-ethereum/accounts/abi/bind"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/core/types"
//	"github.com/ethereum/go-ethereum/crypto"
//	"github.com/ethereum/go-ethereum/ethclient"
//	"github.com/uptrace/bun"
//	"golang.org/x/sync/semaphore"
//	"math"
//	"math/big"
//	"strconv"
//	"time"
//)
//
//type TrackingEvent struct {
//	Wallet   common.Address
//	Contract common.Address
//}
//
//type Event struct {
//	Contract  string
//	Token     string
//	EventName string
//	From      string
//	To        string
//	RawAmount *big.Int
//	Amount    float64
//	Status    string
//	Block     int64
//	Time      time.Time
//	Hash      string
//}
//
//func NewEventTracking(wallet, contract string) (*TrackingEvent, error) {
//	walletAddress := common.HexToAddress(wallet)
//	contractAddress := common.HexToAddress(contract)
//	//
//	//bytecode, err := client.CodeAt(context.Background(), walletAddress, nil) // nil is latest block
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//isWallet := len(bytecode) > 0
//	//if isWallet == false {
//	//	return nil, fmt.Errorf("invalid wallet address ")
//	//}
//	//
//	//bytecode, err = client.CodeAt(context.Background(), contractAddress, nil) // nil is latest block
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//isContract := len(bytecode) > 0
//	//if isContract == false {
//	//	return nil, fmt.Errorf("invalid contract address ")
//	//}
//
//	tracking := &TrackingEvent{
//		Wallet:   walletAddress,
//		Contract: contractAddress,
//	}
//
//	return tracking, nil
//}
//
//func (tracking *TrackingEvent) GetEventFromBlockNumber(db *bun.DB, client *ethclient.Client, number *big.Int, msg chan string) {
//	sem := semaphore.NewWeighted(2)
//	for i := number.Int64(); ; i++ {
//		blockNumber := big.NewInt(i)
//
//		err := sem.Acquire(context.Background(), 1)
//		if err != nil {
//			fmt.Println(err)
//		}
//
//		go func() {
//			event, err := tracking.GetEventByBlockNumber(client, blockNumber)
//			if err != nil {
//				fmt.Println(blockNumber, err)
//				sem.Release(1)
//				return
//			}
//
//			err = InsertEventToDb(db, event)
//			if err != nil {
//				fmt.Println(blockNumber, err)
//			}
//			SendEventToBot(msg, event)
//			sem.Release(1)
//		}()
//	}
//}
//
//func SendEventToBot(msg chan string, data []Event) {
//	for _, tx := range data {
//		message := fmt.Sprintf("Token: %s \nName: %s \nFrom: %s \nTo: %s \nAmount: %s \nHash: %s \nStatus: %s \nTime: %s \nBlock: %s",
//			tx.Token,
//			tx.EventName,
//			tx.From,
//			tx.To,
//			fmt.Sprintf("%g", tx.Amount),
//			tx.Hash,
//			tx.Status,
//			tx.Time.String(),
//			strconv.FormatInt(tx.Block, 10),
//		)
//		msg <- message
//	}
//}
//
//func InsertEventToDb(db *bun.DB, data []Event) error {
//	for _, tx := range data {
//		event := tables.Event{
//			Token:     tx.Token,
//			From:      tx.From,
//			To:        tx.To,
//			Amount:    tx.Amount,
//			RawAmount: tx.RawAmount,
//			EventName: tx.EventName,
//			Status:    tx.Status,
//			Time:      tx.Time,
//			Block:     tx.Block,
//			Hash:      tx.Hash,
//		}
//
//		_, err := db.NewInsert().
//			Model(&event).
//			Exec(context.Background())
//		if err != nil {
//			return err
//		}
//
//		fmt.Println("Inserting to database ...")
//	}
//
//	return nil
//}
//
//func (tracking *TrackingEvent) GetEventByBlockNumber(client *ethclient.Client, number *big.Int) ([]Event, error) {
//	query := ethereum.FilterQuery{
//		FromBlock: number,
//		ToBlock:   number,
//		Addresses: []common.Address{
//			tracking.Contract,
//		},
//	}
//
//	logs, err := client.FilterLogs(context.Background(), query)
//	if err != nil {
//		return nil, err
//	}
//
//	logTransferSig := []byte("Transfer(address,address,uint256)")
//	//logWithdrawSig := []byte("Withdrawal(address,uint256)")
//
//	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
//	//logWithdrawSigHash := crypto.Keccak256Hash(logWithdrawSig)
//	var event []Event
//	for _, vLog := range logs {
//		switch vLog.Topics[0].Hex() {
//		case logTransferSigHash.Hex():
//			err = tracking.GetTransferEventData(client, vLog, &event)
//			if err != nil {
//				fmt.Println(err)
//			}
//
//			//case logWithdrawSigHash.Hex():
//			//	err = tracking.GetWithdrawEventData(client, vLog, &event)
//			//	if err != nil {
//			//		fmt.Println(err)
//			//	}
//		}
//
//	}
//	time.Sleep(1000 * time.Millisecond)
//	return event, nil
//}
//
//func (tracking *TrackingEvent) GetTransferEventData(client *ethclient.Client, vLog types.Log, event *[]Event) error {
//	data := new(big.Int)
//	fromAddress := common.HexToAddress(vLog.Topics[1].Hex())
//	toAddress := common.HexToAddress(vLog.Topics[2].Hex())
//
//	status := tracking.GetTransferStatus(fromAddress, toAddress)
//	if status == "null" {
//		return nil
//	}
//
//	amount := data.SetBytes(vLog.Data)
//	blockTime, err := GetBlockTimestamp(client, vLog.BlockNumber)
//	if err != nil {
//		return err
//	}
//
//	token, err := GetTokenName(client, vLog.Address)
//	if err != nil {
//		return err
//	}
//
//	*event = append(*event, Event{
//		Contract:  vLog.Address.String(),
//		Token:     *token,
//		EventName: "Transfer",
//		From:      fromAddress.String(),
//		To:        toAddress.String(),
//		RawAmount: amount,
//		Amount:    AmountFloatPrecision(amount.Int64(), 6),
//		Status:    status,
//		Block:     int64(vLog.BlockNumber),
//		Time:      blockTime,
//		Hash:      vLog.TxHash.String(),
//	})
//
//	return nil
//}
//
////func (tracking *TrackingEvent) GetWithdrawEventData(client *ethclient.Client, vLog types.Log, event *[]Event) error {
////	data := new(big.Int)
////	fromAddress := common.HexToAddress(vLog.Topics[1].Hex())
////	toAddress := vLog.Address
////
////	status := tracking.GetTransferStatus(fromAddress, toAddress)
////	if status == "null" {
////		return nil
////	}
////
////	amount := data.SetBytes(vLog.Data)
////	blockTime, err := GetBlockTimestamp(client, vLog.BlockNumber)
////	if err != nil {
////		return err
////	}
////
////	*event = append(*event, Event{
////		Contract:  vLog.Address.String(),
////		Name:      "Withdraw",
////		From:      fmt.Sprintf("%s", fromAddress),
////		To:        fmt.Sprintf("%s", toAddress),
////		RawAmount: amount,
////		Amount:    AmountFloatPrecision(amount.Int64(), 6),
////		Status:    status,
////		Block:     int64(vLog.BlockNumber),
////		Time:      blockTime,
////		Hash:      vLog.TxHash.String(),
////	})
////
////	return nil
////}
//
//func (tracking *TrackingEvent) GetTransferStatus(fromAddress common.Address, toAddress common.Address) string {
//	if fromAddress == tracking.Wallet {
//		return "out"
//	}
//
//	if toAddress == tracking.Wallet {
//		return "in"
//	}
//
//	return "null"
//}
//
//func GetBlockTimestamp(client *ethclient.Client, number uint64) (time.Time, error) {
//	block, err := client.BlockByNumber(context.Background(), new(big.Int).SetUint64(number))
//	if err != nil {
//		return time.Time{}, err
//	}
//
//	blockTime := int64(block.Time())
//	return time.Unix(blockTime, 0), nil
//}
//
//func AmountFloatPrecision(num int64, precision int) float64 {
//	n := float64(num) / math.Pow10(18)
//	p := math.Pow10(precision)
//	value := float64(int(n*p)) / p
//	return value
//}
//
//func GetTokenName(client *ethclient.Client, address common.Address) (*string, error) {
//	instance, err := NewToken(address, client)
//	if err != nil {
//		return nil, err
//	}
//
//	name, err := instance.Name(&bind.CallOpts{})
//	if err != nil {
//		return nil, err
//	}
//
//	return &name, nil
//}
