package transaction

import (
	"fmt"
)

func (tracking *TrackingTransaction) SendTransactionToBot(msg chan string, data []Transaction) {
	for _, tx := range data {
		if tx.Amount > 0 {
			message := tracking.MessageFormToBot(tx.From, tx.To, "BNB", "0x0000000000000000000000000000000000000000",
				fmt.Sprintf("%g", tx.Amount),
				tx.Hash)
			msg <- *message
		}

		tracking.GetEventStringToBot(msg, tx.Event)
	}
}

func (tracking *TrackingTransaction) MessageFormToBot(from, to, token, tokenAddress, amount, hash string) *string {
	if from == tracking.Address.String() {
		msg := fmt.Sprintf("Address: %s\nSent: %s %s\nTo: %s\nToken Address: %s\nTx Hash: %s",
			from,
			amount,
			token,
			to,
			tokenAddress,
			hash,
		)
		return &msg
	}

	msg := fmt.Sprintf("Address: %s\nReceiver: %s %s\nFrom: %s\nToken Address: %s\nTx Hash: %s",
		to,
		amount,
		token,
		to,
		tokenAddress,
		hash,
	)
	return &msg
}

func (tracking *TrackingTransaction) GetEventStringToBot(msg chan string, event []Event) {
	if event == nil {
		return
	}

	for _, ev := range event {
		message := tracking.MessageFormToBot(ev.From, ev.To, ev.Token, ev.TokenAddress, fmt.Sprintf("%g", ev.Amount), ev.Hash)
		msg <- *message
	}
	return
}
