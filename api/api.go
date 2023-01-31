package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"log"
	"os"
)

type GinEngine struct {
	g *gin.Engine
}

var db *bun.DB

func NewGin(database *bun.DB) *GinEngine {
	db = database
	return &GinEngine{g: gin.New()}
}

func (gin *GinEngine) Run() {
	gin.Route()
	err := gin.g.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		log.Fatal(err)
	}
}

func (gin *GinEngine) Route() {
	gin.g.GET("/transactions", GetTransactions)
	gin.g.GET("/transactions/:hash", GetTransactionsByHash)
}
