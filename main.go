package main

import (
	"ether/api"
	"ether/bot"
	"ether/database"
	"ether/tables"
	"ether/transaction"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
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

	bot, err := bot.NewBot(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	apiGin := api.NewGin(db)

	messageTx := make(chan string)

	trackingTx, err := transaction.NewTxTracking(os.Getenv("WALLET_ADDRESS"))
	if err != nil {
		log.Fatal(err)
	}
	go apiGin.Run()
	go func() {
		err = bot.Handler()
		if err != nil {
			fmt.Println(err)
		}
	}()

	go trackingTx.GetTransactionFromBlockNumber(db, client, big.NewInt(fromBlock), messageTx)

	for {
		select {
		case msg := <-messageTx:
			err = SendMsgToUsers(bot, bot.GetTxUserList(), msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func SendMsgToUsers(bot *bot.Bot, users []*tele.User, msg string) error {
	for _, user := range users {
		err := bot.Send(user, msg)
		if err != nil {
			return err
		}
	}
	return nil
}
