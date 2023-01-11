package main

import (
	"ether/database"
	"ether/tables"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"os"
	"strconv"
)

func main() {
	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	err = tables.Initial(db)
	if err != nil {
		log.Fatal(err)
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client, err := ethclient.Dial(os.Getenv("RPC"))
	if err != nil {
		log.Fatal(err)
	}

	fromBlock, err := strconv.ParseInt(os.Getenv("FROM_BLOCK"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	trackingEvent := TrackingEvent{
		wallet:   common.HexToAddress(os.Getenv("WALLET_ADDRESS")),
		contract: common.HexToAddress(os.Getenv("CONTRACT_ADDRESS")),
	}

	trackingEvent.GetEventFromBlockNumber(db, client, big.NewInt(fromBlock))

	//trackingTx := TrackingTransaction{
	//	address: os.Getenv("WALLET_ADDRESS"),
	//}
	//trackingTx.GetTransactionFromBlockNumber(db, client, big.NewInt(fromBlock))
}
