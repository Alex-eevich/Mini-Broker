package tbank

import (
	"context"
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

	err := c.do(
		"POST",
		"tinkoff.public.invest.api.contract.v1.InstrumentsService/Shares",
		req,
		&instrument,
	)

	if err != nil {
		fmt.Println("Shares: Ошибка получения списка ticker'ов", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	pool, err := db.ConnectDB()
	if err != nil {
		return err
	} else {
		log.Println("AddTicker: Подключение к БД успешно!")
	}
	db_conn := DB_connect{
		DB: pool,
	}

	conn, err := db_conn.DB.Begin(context.Background())
	if err != nil {
		log.Println("AddTicker: Ошибка подключения к БД")
		return err
	} else {
		log.Println("AddTicker: открыто соединение к БД")
	}
	defer conn.Rollback(context.Background())

	for _, i := range instrument.Instruments {
		_, err = conn.Exec(context.Background(), `
		insert into tickers (ticker, figi, country, create_time)
		values ($1, $2, $3, now())
		`, i.Ticker, i.Figi, i.CountryOfRiskName)
		if err != nil {
			return err
		} else {
			log.Println("AddTicker: Ticker ", i.Ticker, " добавлен в БД")
		}
	}
	conn.Commit(context.Background())

	return nil
}
