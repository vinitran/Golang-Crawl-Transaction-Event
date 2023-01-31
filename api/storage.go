package api

import (
	"context"
	"ether/tables"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"net/http"
)

func GetTransactionsByHash(c *gin.Context) {
	reponseData := new([]tables.Transaction)
	pageFilter := new(PageFilter)
	hash := c.Param("hash")

	err := pageFilter.Check(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		fmt.Println(err)
		return
	}

	offset := (pageFilter.Page - 1) * pageFilter.Size
	query := db.NewSelect().Model(reponseData).
		Where("hash = ?", hash).
		Limit(pageFilter.Size).
		Offset(offset)

	err = query.Scan(context.Background())
	if err != nil {
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		fmt.Println(err)
		return
	}

	c.JSON(http.StatusOK, render.JSON{Data: reponseData})
	return
}

func GetTransactions(c *gin.Context) {
	reponseData := new([]tables.Transaction)
	pageFilter := new(PageFilter)

	err := pageFilter.Check(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		fmt.Println(err)
		return
	}

	offset := (pageFilter.Page - 1) * pageFilter.Size
	query := db.NewSelect().Model(reponseData).
		Limit(pageFilter.Size).
		Offset(offset)

	err = SearchByToken(c, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		fmt.Println(err)
		return
	}

	err = SearchByBlock(c, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		fmt.Println(err)
		return
	}

	err = SearchByAmount(c, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		fmt.Println(err)
		return
	}

	err = SearchByTime(c, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		fmt.Println(err)
		return
	}

	err = SearchByStatus(c, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		fmt.Println(err)
		return
	}

	err = SortByAmount(c, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		fmt.Println(err)
		return
	}

	err = query.Scan(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	c.JSON(http.StatusOK, render.JSON{Data: reponseData})
	return
}
