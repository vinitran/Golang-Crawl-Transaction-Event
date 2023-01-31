package tables

import (
	"context"
	"github.com/uptrace/bun"
	"time"
)

//type Transaction struct {
//	bun.BaseModel `bun:"table:transaction"`
//	Id            int       `bun:"id,pk,autoincrement"`
//	From          string    `bun:"from,notnull"`
//	To            string    `bun:"to"`
//	RawAmount     string    `bun:"raw_amount,notnull"`
//	Amount        float64   `bun:"amount,notnull"`
//	Hash          string    `bun:"hash,notnull"`
//	Status        string    `bun:"status,notnull"`
//	Time          time.Time `bun:"time,notnull"`
//	Block         int64     `bun:"block,notnull"`
//	Event         []*Event  `bun:"rel:has-many,join:hash=hash"`
//}
//
//type Event struct {
//	bun.BaseModel `bun:"table:event"`
//	Id            int     `bun:"id,pk,autoincrement"`
//	From          string  `bun:"from,notnull"`
//	To            string  `bun:"to"`
//	Token         string  `bun:"token"`
//	RawAmount     string  `bun:"raw_amount,notnull"`
//	Amount        float64 `bun:"amount,notnull"`
//	Hash          string  `bun:"hash,notnull"`
//}

type Transaction struct {
	bun.BaseModel `bun:"table:transaction"`
	Id            int       `bun:"id,pk,autoincrement"`
	Token         string    `bun:"token,notnull"`
	From          string    `bun:"from,notnull"`
	To            string    `bun:"to"`
	RawAmount     string    `bun:"raw_amount,notnull"`
	Amount        float64   `bun:"amount,notnull"`
	Hash          string    `bun:"hash,notnull"`
	Status        string    `bun:"status,notnull"`
	Time          time.Time `bun:"time,notnull"`
	Block         int64     `bun:"block,notnull"`
	TokenAddress  string    `bun:"token_address,notnull"`
}

func Initial(db *bun.DB) error {
	_, err := db.NewCreateTable().
		Model((*Transaction)(nil)).
		IfNotExists().
		Exec(context.Background())
	if err != nil {
		return err
	}

	//
	//_, err = db.NewCreateTable().
	//	Model((*Event)(nil)).
	//	IfNotExists().
	//	Exec(context.Background())
	//if err != nil {
	//	return err
	//}

	return nil
}
