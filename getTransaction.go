package main

import (
	"context"
	"ether/tables"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/uptrace/bun"
	"golang.org/x/sync/semaphore"
	"math/big"
	"time"
)

type TrackingTransaction struct {
	address string
}

type Transaction struct {
	From   string
	To     string
	Amount string
	Hash   string
	Status string
	Time   time.Time
	Block  int64
}

func (tracking TrackingTransaction) GetTransactionFromBlockNumber(db *bun.DB, client *ethclient.Client, number *big.Int) {
	sem := semaphore.NewWeighted(1)
	for i := number.Int64(); ; i++ {
		blockNumber := big.NewInt(i)
		sem.Acquire(context.Background(), 1)
		go func() {
			data, err := tracking.GetTransactionByBlockNumber(client, blockNumber)
			if err != nil {
				fmt.Println(blockNumber, err)
			}

			err = InsertTransactionToDb(db, data)
			if err != nil {
				fmt.Println(blockNumber, err)
			}

			sem.Release(1)
		}()
	}
}

func InsertTransactionToDb(db *bun.DB, data []Transaction) error {
	for _, tx := range data {
		transaction := tables.Transaction{
			From:   tx.From,
			To:     tx.To,
			Amount: tx.Amount,
			Hash:   tx.Hash,
			Status: tx.Status,
			Time:   tx.Time,
			Block:  tx.Block,
		}

		_, err := db.NewInsert().
			Model(&transaction).
			Exec(context.Background())
		if err != nil {
			return err
		}

		fmt.Println("Inserting to database ...")
	}

	return nil
}

func (tracking TrackingTransaction) GetTransactionByBlockNumber(client *ethclient.Client, number *big.Int) ([]Transaction, error) {
	lastestBlockNumber, err := GetLatestBlockNumber(client)
	if err != nil {
		return nil, err
	}

	if number.Int64() > lastestBlockNumber.Int64() {
		timeDelay := (number.Int64() - lastestBlockNumber.Int64()) * 3
		time.Sleep(time.Duration(timeDelay) * time.Second)
	}

	block, err := client.BlockByNumber(context.Background(), number)
	if err != nil {
		return nil, err
	}

	var data []Transaction
	for _, tx := range block.Transactions() {
		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			return nil, err
		}

		msg, err := tx.AsMessage(types.NewEIP155Signer(chainID), tx.GasPrice())
		if err != nil {
			return nil, err
		}

		isInTransaction, status := IsInTransaction(tracking.address, msg, tx)
		if isInTransaction == false {
			continue
		}

		recipient := GetRecipient(tx)
		timestamp := time.Unix(int64(block.Time()), 0)
		data = append(data, Transaction{
			From:   msg.From().Hex(),
			To:     recipient,
			Amount: tx.Value().String(),
			Hash:   tx.Hash().Hex(),
			Status: status,
			Time:   timestamp,
			Block:  block.Number().Int64(),
		})
	}
	return data, nil
}

func GetLatestBlockNumber(client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return header.Number, nil
}

//func GetBalance(account common.Address, client *ethclient.Client, number *big.Int) (*big.Int, error) {
//	balance, err := client.BalanceAt(context.Background(), account, number)
//	if err != nil {
//		return nil, err
//	}
//
//	return balance, nil
//}

func IsInTransaction(address string, msg types.Message, tx *types.Transaction) (bool, string) {
	if msg.From().Hex() == address {
		return true, "out"
	}

	if tx.To() == nil {
		return false, "null"
	}

	if tx.To().Hex() == address {
		return true, "in"
	}
	return false, "null"
}

func GetRecipient(tx *types.Transaction) string {
	if tx.To() == nil {
		return "null"
	}

	return tx.To().Hex()
}
