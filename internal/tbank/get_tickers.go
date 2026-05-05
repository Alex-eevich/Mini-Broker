package tbank

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mini-broker/internal/db"
	"net/http"
)

type request struct {
	instrumentStatus string `json:"instrumentStatus"`
}

type instruments struct {
	Instruments []struct {
		Figi              string `json:"figi"`
		Ticker            string `json:"ticker"`
		CountryOfRiskName string `json:"countryOfRiskName"`
	} `json:"instruments"`
}

func (c *Client) Shares(w http.ResponseWriter, _ *http.Request) {
	var instrument instruments

	req := request{
		instrumentStatus: "INSTRUMENT_STATUS_BASE",
	}

	postErr := c.do(
		"POST",
		"tinkoff.public.invest.api.contract.v1.InstrumentsService/Shares",
		req,
		&instrument,
	)

	if postErr != nil {
		fmt.Println("Shares: Ошибка получения списка ticker'ов", postErr)
		http.Error(w, postErr.Error(), http.StatusInternalServerError)
		return
	}

	addTickerErr := AddTicker(instrument)
	if addTickerErr == nil {
		log.Println("AddTicker: Все tickers добавлены в БД. Справочник обновлен.")
	}

	w.WriteHeader(http.StatusOK)
	return
}

func AddTicker(instrument instruments) error {
	pool, connErr := db.ConnectDB()
	if connErr != nil {
		return connErr
	} else {
		log.Println("AddTicker: Подключение к БД успешно!")
	}
	dbConn := DB_connect{
		DB: pool,
	}

	conn, dbErr := dbConn.DB.Begin(context.Background())
	if dbErr != nil {
		log.Println("AddTicker: Ошибка подключения к БД")
		return dbErr
	} else {
		log.Println("AddTicker: открыто соединение к БД")
	}
	defer conn.Rollback(context.Background())

	for _, i := range instrument.Instruments {
		_, insertErr := conn.Exec(context.Background(), `
		insert into tickers (ticker, figi, country, create_time)
		values ($1, $2, $3, now())
		`, i.Ticker, i.Figi, i.CountryOfRiskName)
		if insertErr != nil {
			return insertErr
		} else {
			log.Println("AddTicker: Ticker ", i.Ticker, " добавлен в БД")
		}
	}
	conn.Commit(context.Background())

	return nil
}

func (c *Client) GetTickers() ([]string, error) {
	pool, _ := db.ConnectDB()
	dbConn := DB_connect{
		DB: pool,
	}

	conn, connErr := dbConn.DB.Begin(context.Background())
	if connErr != nil {
		return nil, connErr
	}

	tickers, tickErr := conn.Query(context.Background(),
		`select ticker from tickers`)
	defer tickers.Close()
	if tickErr != nil {
		return nil, tickErr
	}
	var listTickers []string
	for tickers.Next() {
		var t string
		err := tickers.Scan(&t)
		if err != nil {
			return nil, err
		}
		listTickers = append(listTickers, t)
	}

	return listTickers, nil
}

func (c *Client) ListTickers(w http.ResponseWriter, _ *http.Request) {
	tickers, getErr := c.GetTickers()
	if getErr != nil {
		http.Error(w, getErr.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickers)
}
