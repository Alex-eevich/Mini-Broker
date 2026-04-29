package tbank

import (
	"context"
	"encoding/json"
	"log"
	"mini-broker/internal/db"
	"mini-broker/internal/user_data"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

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
	Ticker    string
	Figi      string `json:"figi"`
	EventFrom string `json:"from"`
	EventTo   string `json:"to"`
	Interval  string `json:"interval"`
}

type MoneyValue struct {
	Units string `json:"units"`
	Nano  int32  `json:"nano"`
}

type Candle struct {
	Time   time.Time  `json:"time"`
	Open   MoneyValue `json:"open"`
	High   MoneyValue `json:"high"`
	Low    MoneyValue `json:"low"`
	Close  MoneyValue `json:"close"`
	Volume int64      `json:"volume,string"`
}

type ResponseItemAPI struct {
	Ticker  string   `json:"ticker"`
	Candles []Candle `json:"candles"`
}

type CandleRes struct {
	Time  int64   `json:"time"`
	Close float64 `json:"close"`
}

type ResponseItem struct {
	Ticker    string      `json:"ticker"`
	CandleRes []CandleRes `json:"candles"`
}

type Results struct {
	Results []ResponseItem `json:"results"`
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
	var trade_token string
	query := `
		select token from user_tradetoken
		where user_id = $1;`
	selectErr := pool.QueryRow(context.Background(), query, user_id).Scan(&trade_token)
	if selectErr != nil {
		log.Println(selectErr.Error())
	}

	log.Println(trade_token)
	log.Println(filt)
	reqJson := c.RebuildJsonFilter(filt, pool)
	for _, req := range reqJson.Requests {
		log.Println(req)
	}
	candles, err := c.TbankGetCandles(reqJson)
	resultsCandle := BuildResponse(candles)
	log.Println(resultsCandle)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	errdecode := json.NewEncoder(w).Encode(resultsCandle)
	if errdecode != nil {
		log.Println("encode error:", errdecode)
	} else {
		log.Println("Отдали ответ на /getGraph/")
	}
	log.Println("DONE RESPONSE")
	return
}

func (c *Client) RebuildJsonFilter(filter Filter, pool *pgxpool.Pool) Request {
	tickers := []string{
		filter.Ticker1,
		filter.Ticker2,
		filter.Ticker3,
		filter.Ticker4,
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

		err := pool.QueryRow(
			context.Background(),
			`select figi from tickers where ticker = $1`,
			ticker,
		).Scan(&figi)

		if err != nil {
			log.Println(err)
			continue
		}

		req.Requests = append(req.Requests, Item{
			Ticker:    ticker,
			Figi:      figi,
			EventFrom: eventFrom,
			EventTo:   eventTo,
			Interval:  intervalMap[filter.Interval],
		})
	}
	return req
}

func (c *Client) TbankGetCandles(request Request) ([]ResponseItemAPI, error) {
	var wg sync.WaitGroup
	results := make([]ResponseItemAPI, len(request.Requests))
	errors := make([]error, len(request.Requests))

	wg.Add(len(request.Requests))

	for i, req := range request.Requests {
		go func(i int, req Item) {
			defer wg.Done()

			res, errGoroutines := c.GorouteGetCandles(req)
			if errGoroutines != nil {
				errors[i] = errGoroutines
				return
			}
			results[i] = res

		}(i, req)
	}
	wg.Wait()
	return results, nil
}

func (c *Client) GorouteGetCandles(req Item) (ResponseItemAPI, error) {

	log.Println(req.Ticker)

	var response ResponseItemAPI
	errApi := c.do(
		"POST",
		"tinkoff.public.invest.api.contract.v1.MarketDataService/GetCandles",
		req,
		&response,
	)
	if errApi != nil {
		return ResponseItemAPI{}, errApi
	} else {
		log.Println("GorouteGetCandles: Запрос на Candles отправлен успешно!")
	}
	response.Ticker = req.Ticker
	return response, nil
}

func BuildResponse(data []ResponseItemAPI) Results {
	results := Results{}

	for _, d := range data {
		item := ResponseItem{
			Ticker: d.Ticker,
		}
		for _, cd := range d.Candles {

			closecandles := parseMoney(cd.Close)
			item.CandleRes = append(item.CandleRes, CandleRes{
				Time:  cd.Time.Unix(),
				Close: closecandles,
			})
		}

		results.Results = append(results.Results, item)
	}
	return results
}

func parseMoney(m MoneyValue) float64 {
	units, _ := strconv.ParseFloat(m.Units, 64)
	nano := float64(m.Nano) / 1e9
	return units + nano
}
