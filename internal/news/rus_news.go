package news

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

func RussianNews(w http.ResponseWriter, _ *http.Request) {

	parser := gofeed.NewParser()
	var newsRus []News
	log.Println("RussianNews: Парсинг новостей.")
	feed, feedErr := parser.ParseURL("https://www.interfax.ru/rss.asp")
	log.Println("RussianNews: Результаты:")
	if feedErr != nil {
		http.Error(w, feedErr.Error(), http.StatusInternalServerError)
		log.Println("RussianNews: Parse Error")
		return
	}

	for _, item := range feed.Items {

		image := ""
		if item.Image != nil {
			image = item.Image.URL
		}

		category := ""
		if len(item.Categories) > 0 {
			category = item.Categories[0]
		}

		published := item.Published
		dateParse, dateParseErr := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", published)
		if dateParseErr != nil {
			log.Println("RussianNews: dateParseErr.", dateParseErr)
		}
		publishTime := dateParse.Format("Mon, 02 Jan 2006 15:04")

		oneNews := News{
			Title:       item.Title,
			Description: item.Description,
			URL:         item.Link,
			Image:       image,
			PublishedAt: publishTime,
			Categories:  category,
		}
		newsRus = append(newsRus, oneNews)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newsRus)
	return
}
