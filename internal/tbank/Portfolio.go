package tbank

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Account struct {
	Accounts []struct {
		Id string `json:"id"`
	} `json:"accounts"`
}

type SandboxPortfolio struct {
	AccountId string `json:"accountId"`
	Currency  string `json:"currency"`
}

type DB_connect struct {
	DB *pgxpool.Pool
}

func (c *Client) GetListAccounts() (string, error) {
	var account Account

	err := c.do(
		"POST",
		"tinkoff.public.invest.api.contract.v1.SandboxService/GetSandboxAccounts",
		map[string]string{},
		&account,
	)
	if err != nil {
		return "", err
	}

	var accounts []string
	for _, account := range account.Accounts {
		accounts = append(accounts, account.Id)
	}
	log.Printf("Accounts: %v", accounts)
	log.Printf("Будем работать с : %v", accounts[0])
	return accounts[0], nil
}

func (c *Client) OpenSandboxAccount() (string, error) {
	var resp struct {
		AccountId string `json:"accountId"`
	}

	err := c.do(
		"POST",
		"tinkoff.public.invest.api.contract.v1.SandboxService/OpenSandboxAccount",
		map[string]string{},
		&resp,
	)
	if err != nil {
		log.Println("AddAccount: Не удалось создать торговый счет для пользователя!", err)
		return "", err
	}
	return resp.AccountId, nil
}

func (dbc *DB_connect) AddTradeAccountDB(accountId string, user_id int) error {
	tx, err := dbc.DB.Begin(context.Background())
	if err != nil {
		log.Println("AddTradeAccountDB: Ошибка подключения к БД для добавления TradeAccount")
		return err
	}
	defer tx.Rollback(context.Background())

	var token_id int
	querytt := `
		select id from user_tradetoken
		where user_id = $1;`
	selectttErr := tx.QueryRow(context.Background(), querytt, user_id).Scan(&token_id)
	if selectttErr != nil {
		log.Println("AddTradeAccountDB: действующий токен не найден")
	} else {
		log.Println("Добавляем trade_account к токену ", token_id)
	}

	queryta := `
		select id from trade_account
		where user_id = $1;`
	selecttaErr := tx.QueryRow(context.Background(), queryta, user_id).Scan(&token_id)
	if selecttaErr != nil {
		log.Println("AddTradeAccountDB: действующий торговый счет уже существует в БД")
	}

	_, err = tx.Exec(context.Background(), `
		insert into trade_account (user_id, user_token_id, number, create_time)
		values ($1, $2, $3, now())
		`, user_id, token_id, accountId)

	err = tx.Commit(context.Background())
	if err != nil {
		log.Println("AddTradeAccountDB: commit insert trade_account error", err.Error())
		return err
	}

	if err != nil {
		return err
	} else {
		log.Println("AddTradeAccountDB: trade_account пользователя был добавлен в БД")
		return nil
	}
}
