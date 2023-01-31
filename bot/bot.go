package bot

import (
	tele "gopkg.in/telebot.v3"
	"log"
	"time"
)

type Bot struct {
	b *tele.Bot
}

var eventUserList []*tele.User
var transactionUserList []*tele.User

func NewBot(token string) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	tmpBot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	bot := &Bot{
		b: tmpBot,
	}
	return bot, nil
}

func (bot *Bot) GetEventUserList() []*tele.User {
	return eventUserList
}

func (bot *Bot) GetTxUserList() []*tele.User {
	return transactionUserList
}

func (bot *Bot) Send(user *tele.User, msg string) error {
	_, err := bot.b.Send(user, msg)
	if err != nil {
		return err
	}

	return nil
}

func (bot *Bot) Handler() error {
	bot.b.Handle("/handlerEvent", func(c tele.Context) error {
		eventUserList = append(eventUserList, c.Sender())
		return c.Send("Tracking...!")
	})

	bot.b.Handle("/stopHandlerEvent", func(c tele.Context) error {
		eventUserList = nil
		return c.Send("Stop Tracking!")
	})

	bot.b.Handle("/handlerTx", func(c tele.Context) error {
		transactionUserList = append(transactionUserList, c.Sender())
		return c.Send("Tracking...!")
	})

	bot.b.Handle("/stopHandlerTx", func(c tele.Context) error {
		transactionUserList = nil
		return c.Send("Stop Tracking!")
	})

	bot.b.Start()
	return nil
}
