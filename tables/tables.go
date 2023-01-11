package tables

import (
	"context"
	"github.com/uptrace/bun"
	"math/big"
	"time"
)

type Transaction struct {
	bun.BaseModel `bun:"table:transaction"`
	Id            int       `bun:"id,pk,autoincrement"`
	From          string    `bun:"from,notnull"`
	To            string    `bun:"to"`
	Amount        string    `bun:"amount,notnull"`
	Hash          string    `bun:"hash,notnull"`
	Status        string    `bun:"status,notnull"`
	Time          time.Time `bun:"time,notnull"`
	Block         int64     `bun:"block,notnull"`
}

type Event struct {
	bun.BaseModel `bun:"table:event"`
	Id            int       `bun:"id,pk,autoincrement"`
	From          string    `bun:"from,notnull"`
	To            string    `bun:"to"`
	Amount        *big.Int  `bun:"amount,notnull"`
	Name          string    `bun:"name,notnull"`
	Status        string    `bun:"status,notnull"`
	Time          time.Time `bun:"time,notnull"`
	Block         int64     `bun:"block,notnull"`
	Hash          string    `bun:"hash,notnull"`
}

func Initial(db *bun.DB) error {
	_, err := db.NewCreateTable().
		Model((*Transaction)(nil)).
		IfNotExists().
		Exec(context.Background())
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		Model((*Event)(nil)).
		IfNotExists().
		Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}
