package transaction

import (
	"context"
	"ether/tables"
	"fmt"
	"github.com/uptrace/bun"
)

func (tracking *TrackingTransaction) InsertTransactionToDb(db *bun.DB, data []Transaction) error {
	for _, tx := range data {
		if tx.Amount > 0 {
			transaction := tables.Transaction{
				From:         tx.From,
				To:           tx.To,
				RawAmount:    tx.RawAmount,
				Amount:       tx.Amount,
				Token:        "BNB",
				Hash:         tx.Hash,
				Status:       tx.Status,
				Time:         tx.Time,
				Block:        tx.Block,
				TokenAddress: "0x0000000000000000000000000000000000000000",
			}

			_, err := db.NewInsert().
				Model(&transaction).
				Exec(context.Background())
			if err != nil {
				return err
			}
			fmt.Println("Inserting to database ...")
		}

		for _, ev := range tx.Event {
			if ev.Amount <= 0 {
				continue
			}

			transaction := tables.Transaction{
				From:         ev.From,
				To:           ev.To,
				RawAmount:    ev.RawAmount,
				Amount:       ev.Amount,
				Token:        ev.Token,
				Hash:         ev.Hash,
				Status:       ev.Status,
				Time:         tx.Time,
				Block:        tx.Block,
				TokenAddress: ev.TokenAddress,
			}

			_, err := db.NewInsert().
				Model(&transaction).
				Exec(context.Background())
			if err != nil {
				return err
			}
			fmt.Println("Inserting to database ...")
		}
	}

	return nil
}
