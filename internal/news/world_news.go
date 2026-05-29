package news

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func WorldNews(w http.ResponseWriter, _ *http.Request) {
	url := os.Getenv("FINNHUB_URL_TOKEN")
	resp, respErr := http.Get(url)
	if respErr != nil {
		http.Error(w, respErr.Error(), http.StatusInternalServerError)
		log.Println("WorldNews: Не удалось подключиться к Finnhub.io")
		return
	}
	defer resp.Body.Close()

	var news []FinnhubNews
	err := json.NewDecoder(resp.Body).Decode(&news)
	if err != nil {
		log.Println("WorldNews: не смог прочитать json")
	}
	var response []News
	for _, item := range news {
		PublishedTime := time.Unix(item.Datetime, 0).Format("Mon, 02 Jan 2006 15:04")
		news := News{
			Title:       item.Headline,
			Description: item.Summary,
			URL:         item.URL,
			Image:       item.Image,
			PublishedAt: PublishedTime,
			Categories:  item.Source,
		}
		response = append(response, news)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return
}
