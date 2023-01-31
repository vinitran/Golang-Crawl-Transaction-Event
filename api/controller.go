package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"strconv"
	"time"
)

type PageFilter struct {
	Page int
	Size int
}

func SearchByTime(c *gin.Context, query *bun.SelectQuery) error {
	strTimeFrom, isTimeFrom := c.GetQuery("from_time")
	strTimeTo, isTimeTo := c.GetQuery("to_time")
	if !isTimeFrom && !isTimeTo {
		return nil
	}

	if !isTimeFrom {
		timeTo, err := time.Parse(time.RFC3339, strTimeTo)
		if err != nil {
			return fmt.Errorf("invalid time format. Use RFC3339 format")
		}

		query = query.Where("time <= ?", timeTo)
		return nil
	}

	if !isTimeTo {
		timeFrom, err := time.Parse(time.RFC3339, strTimeFrom)
		if err != nil {
			return fmt.Errorf("invalid time format. Use RFC3339 format")
		}

		query = query.Where("time >= ?", timeFrom)
		return nil
	}

	timeTo, err := time.Parse(time.RFC3339, strTimeTo)
	if err != nil {
		return fmt.Errorf("invalid time format. Use RFC3339 format")
	}

	timeFrom, err := time.Parse(time.RFC3339, strTimeFrom)
	if err != nil {
		return fmt.Errorf("invalid time format. Use RFC3339 format")
	}

	if timeFrom.After(timeTo) {
		return fmt.Errorf("invalid time value. to_time must be greater than from_time")
	}

	query = query.Where("time >= ?", timeFrom).Where("time <= ?", timeTo).Order("time ASC")
	return nil
}

func (p *PageFilter) Check(c *gin.Context) error {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("%s", err))
	}
	p.Page = page
	if p.Page <= 0 {
		return fmt.Errorf("error: page must be greater than 0")
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("%s", err))
	}
	p.Size = size
	if p.Size <= 0 {
		return fmt.Errorf("error: size must be greater than 0")
	}

	return nil
}

func SearchByBlock(c *gin.Context, query *bun.SelectQuery) error {
	strBlock, ok := c.GetQuery("block")
	if !ok {
		return nil
	}

	block, err := strconv.Atoi(strBlock)
	if err != nil {
		return fmt.Errorf("error: invalid value for block, only number type")
	}

	if block < 0 {
		return fmt.Errorf("error: invalid value for block, must be grater than or equal to 0")
	}

	query = query.Where("block = ?", block)
	return nil
}

func SearchByToken(c *gin.Context, query *bun.SelectQuery) error {
	token, ok := c.GetQuery("token")
	if !ok {
		return nil
	}

	query = query.Where("token_address = ?", token)
	return nil
}

func SearchByStatus(c *gin.Context, query *bun.SelectQuery) error {
	status, ok := c.GetQuery("status")
	if !ok {
		return nil
	}
	if status != "in" && status != "out" {
		return fmt.Errorf("error: invalid value for status, must be in or out")
	}

	query = query.Where("status = ?", status)
	return nil
}

func SortByAmount(c *gin.Context, query *bun.SelectQuery) error {
	sort, ok := c.GetQuery("sort_by_amount")
	if !ok {
		return nil
	}

	if sort == "desc" {
		query = query.Order("amount DESC")
		return nil
	}

	if sort == "asc" {
		query = query.Order("amount ASC")
		return nil
	}

	return fmt.Errorf("error: invalid value for sort_by_amount, only esc or desc")
}

func SearchByAmount(c *gin.Context, query *bun.SelectQuery) error {
	strAmountFrom, isAmountFrom := c.GetQuery("from_amount")
	strAmountTo, isAmountTo := c.GetQuery("to_amount")

	if !isAmountFrom && !isAmountTo {
		return nil
	}

	if !isAmountFrom {
		amountTo, err := strconv.ParseFloat(strAmountTo, 64)
		if err != nil {
			return fmt.Errorf("error: invalid type value for amount_to, only float type")
		}

		query = query.Where("amount <= ?", amountTo)
		return nil
	}

	if !isAmountTo {
		amountFrom, err := strconv.ParseFloat(strAmountFrom, 64)
		if err != nil {
			return fmt.Errorf("error: invalid type value for amount_from, only float type")
		}

		query = query.Where("amount >= ?", amountFrom)
		return nil
	}

	amountTo, err := strconv.ParseFloat(strAmountTo, 64)
	if err != nil {
		return fmt.Errorf("error: invalid type value for amount_to, only float type")
	}

	amountFrom, err := strconv.ParseFloat(strAmountFrom, 64)
	if err != nil {
		return fmt.Errorf("error: invalid type value for amount_from, only float type")
	}

	if amountFrom >= amountTo {
		return fmt.Errorf("error: invalid value for amount_from and amount_to, amount_to must be greater than amount_from")
	}

	query = query.Where("amount >= ?", amountFrom).Where("amount <= ?", amountTo)
	return nil
}

//func SearchByHash(c *gin.Context, query *bun.SelectQuery) error {
//	hash, ok := c.GetQuery("hash")
//	if !ok {
//		return nil
//	}
//
//	query = query.Where("hash = ?", hash)
//	return nil
//}
