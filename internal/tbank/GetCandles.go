package tbank

import (
	"context"
	"encoding/json"
	"log"
	"mini-broker/internal/db"
	"mini-broker/internal/user_data"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Filter struct {
	Ticker1   string `json:"ticker1"`
	Ticker2   string `json:"ticker2"`
	Ticker3   string `json:"ticker3"`
	Ticker4   string `json:"ticker4"`
	EventFrom string `json:"eventFrom"`
	EventTo   string `json:"eventTo"`
	Interval  string `json:"interval"`
}

type Request struct {
	Requests []Item `json:"requests"`
}

type Item struct {
	Figi      string `json:"figi"`
	EventFrom string `json:"eventFrom"`
	EventTo   string `json:"eventTo"`
	Interval  string `json:"interval"`
}

func (c *Client) GetCandles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		log.Println("Пустой токен!")
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	// Нужно использовать user_id для поиска трейд-токена. Он нужен для создания запроса за кенделсами
	user_id, err := user_data.ParseToken(token)
	if err != nil {
		log.Println(err)
	}
	pool, errConn := db.ConnectDB()
	if errConn != nil {
		log.Println("GetCandles: Ошибка подключения к БД", errConn)
		http.Error(w, errConn.Error(), http.StatusInternalServerError)
		return
	}
	defer pool.Close()
	var filt Filter
	reqErr := json.NewDecoder(r.Body).Decode(&filt)
	if reqErr != nil {
		http.Error(w, reqErr.Error(), http.StatusBadRequest)
		return
	}
	conn, errTa := pool.Begin(context.Background())
	if errTa != nil {
		log.Println(errTa.Error())
	}
	var trade_token string
	query := `
		select token from user_tradetoken
		where user_id = $1;`
	selectErr := conn.QueryRow(context.Background(), query, user_id).Scan(&trade_token)
	if selectErr != nil {
		log.Println(selectErr.Error())
	}

	log.Println(trade_token)
	log.Println(filt)
	reqJson := c.RebuildJsonFilter(filt, pool)
	log.Println(reqJson.Requests[0])
	log.Println(reqJson.Requests[1])
	log.Println(reqJson.Requests[2])
	log.Println(reqJson.Requests[3])
	w.WriteHeader(http.StatusOK)
}

func (c *Client) RebuildJsonFilter(filter Filter, pool *pgxpool.Pool) Request {
	tickers := []string{
		filter.Ticker1,
		filter.Ticker2,
		filter.Ticker3,
		filter.Ticker4,
	}
	conn, err := pool.Begin(context.Background())
	if err != nil {
		log.Println(err)
	}

	var figi string
	eventFrom := filter.EventFrom + ":00.000Z"
	eventTo := filter.EventTo + ":00.000Z"
	var intervalMap = map[string]string{
		"1m":  "CANDLE_INTERVAL_1_MIN",
		"5m":  "CANDLE_INTERVAL_5_MIN",
		"15m": "CANDLE_INTERVAL_15_MIN",
		"1h":  "CANDLE_INTERVAL_HOUR",
		"1d":  "CANDLE_INTERVAL_DAY",
		"1w":  "CANDLE_INTERVAL_WEEK",
	}

	req := Request{}
	for _, ticker := range tickers {
		if ticker == "" {
			continue
		}

		query := `
		select figi from tickers
		where ticker = $1;`
		selectErr := conn.QueryRow(context.Background(), query, ticker).Scan(&figi)
		if selectErr != nil {
			log.Println(selectErr.Error())
		}

		req.Requests = append(req.Requests, Item{
			Figi:      figi,
			EventFrom: eventFrom,
			EventTo:   eventTo,
			Interval:  intervalMap[filter.Interval],
		})
	}
	return req
}
