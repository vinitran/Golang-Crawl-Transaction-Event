package main

import (
	"context"
	"ether/tables"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/uptrace/bun"
	"golang.org/x/sync/semaphore"
	"math/big"
	"time"
)

type TrackingEvent struct {
	wallet   common.Address
	contract common.Address
}

type Event struct {
	Name   string
	From   string
	To     string
	Amount *big.Int
	Status string
	Block  int64
	Time   time.Time
	Hash   string
}

func (tracking TrackingEvent) GetEventFromBlockNumber(db *bun.DB, client *ethclient.Client, number *big.Int) {
	sem := semaphore.NewWeighted(2)
	for i := number.Int64(); ; i++ {
		blockNumber := big.NewInt(i)
		sem.Acquire(context.Background(), 1)
		go func() {
			event, err := tracking.GetEventByBlockNumber(client, blockNumber)
			if err != nil {
				fmt.Println(blockNumber, err)
				sem.Release(1)
				return
			}

			//fmt.Println(blockNumber, event)
			err = InsertEventToDb(db, event)
			if err != nil {
				fmt.Println(blockNumber, err)
			}

			sem.Release(1)
		}()
	}
}

func InsertEventToDb(db *bun.DB, data []Event) error {
	for _, tx := range data {
		event := tables.Event{
			From:   tx.From,
			To:     tx.To,
			Amount: tx.Amount,
			Name:   tx.Name,
			Status: tx.Status,
			Time:   tx.Time,
			Block:  tx.Block,
			Hash:   tx.Hash,
		}

		_, err := db.NewInsert().
			Model(&event).
			Exec(context.Background())
		if err != nil {
			return err
		}

		fmt.Println("Inserting to database ...")
	}

	return nil
}

func (tracking TrackingEvent) GetEventByBlockNumber(client *ethclient.Client, number *big.Int) ([]Event, error) {
	query := ethereum.FilterQuery{
		FromBlock: number,
		ToBlock:   number,
		Addresses: []common.Address{
			tracking.contract,
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}

	logTransferSig := []byte("Transfer(address,address,uint256)")
	logWithdrawSig := []byte("Withdrawal(address,uint256)")

	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
	logWithdrawSigHash := crypto.Keccak256Hash(logWithdrawSig)
	var event []Event
	for _, vLog := range logs {
		switch vLog.Topics[0].Hex() {
		case logTransferSigHash.Hex():
			err = tracking.GetTransferEventData(client, vLog, &event)
			if err != nil {
				fmt.Println(err)
			}

		case logWithdrawSigHash.Hex():
			err = tracking.GetWithdrawEventData(client, vLog, &event)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	time.Sleep(1000 * time.Millisecond)
	return event, nil
}

func (tracking TrackingEvent) GetTransferEventData(client *ethclient.Client, vLog types.Log, event *[]Event) error {
	data := new(big.Int)
	fromAddress := common.HexToAddress(vLog.Topics[1].Hex())
	toAddress := common.HexToAddress(vLog.Topics[2].Hex())

	status := tracking.GetTransferStatus(fromAddress, toAddress)
	if status == "null" {
		return nil
	}

	amount := data.SetBytes(vLog.Data)
	blockTime, hash, err := GetBlockTimestampAndHash(client, vLog.BlockNumber)
	if err != nil {
		return err
	}

	*event = append(*event, Event{
		Name:   "Transfer",
		From:   fmt.Sprintf("%s", fromAddress),
		To:     fmt.Sprintf("%s", toAddress),
		Amount: amount,
		Status: status,
		Block:  int64(vLog.BlockNumber),
		Time:   blockTime,
		Hash:   hash,
	})

	return nil
}

func (tracking TrackingEvent) GetWithdrawEventData(client *ethclient.Client, vLog types.Log, event *[]Event) error {
	data := new(big.Int)
	fromAddress := common.HexToAddress(vLog.Topics[1].Hex())

	amount := data.SetBytes(vLog.Data)
	blockTime, hash, err := GetBlockTimestampAndHash(client, vLog.BlockNumber)
	if err != nil {
		return err
	}

	*event = append(*event, Event{
		Name:   "Withdraw",
		From:   fmt.Sprintf("%s", fromAddress),
		To:     "null",
		Amount: amount,
		Status: "null",
		Block:  int64(vLog.BlockNumber),
		Time:   blockTime,
		Hash:   hash,
	})

	return nil
}

func (tracking TrackingEvent) GetTransferStatus(fromAddress common.Address, toAddress common.Address) string {
	if fromAddress == tracking.wallet {
		return "out"
	}

	if toAddress == tracking.wallet {
		return "in"
	}

	return "null"
}

func GetBlockTimestampAndHash(client *ethclient.Client, number uint64) (time.Time, string, error) {
	block, err := client.BlockByNumber(context.Background(), new(big.Int).SetUint64(number))
	if err != nil {
		return time.Time{}, "null", err
	}

	blockTime := int64(block.Time())
	return time.Unix(blockTime, 0), block.Hash().Hex(), nil
}
