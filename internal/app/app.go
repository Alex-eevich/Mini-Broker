package app

import (
	"log"
	"mini-broker/internal/db"
	"mini-broker/internal/tbank"
	"mini-broker/internal/token"
	"mini-broker/internal/users"
	"net/http"
	"os"
)

func Handler() {
	pool, _ := db.ConnectDB()
	userHandler := &users.Handler{
		DB: pool,
	}
	sender := &users.SMTPEmailSender{}
	emailHandler := &users.EmailService{
		Sender: sender,
		DB:     pool,
	}
	tradeTokenHandler := &token.Token{
		DB: pool,
	}
	tradeAdminToken := os.Getenv("TINKOFF_TOKEN")
	client := tbank.NewClientFromToken(tradeAdminToken)
	http.Handle("/", http.FileServer(http.Dir("./internal/web")))
	http.HandleFunc("/register", users.RegisterUser)
	http.HandleFunc("/auth", users.LoginHandler)
	http.HandleFunc("/adminShares", client.Shares)
	http.HandleFunc("/api/getGraph", client.GetCandles)
	http.HandleFunc("/api/tickers", users.AuthMiddleware(client.ListTickers))
	http.HandleFunc("/verify-email", users.AuthMiddleware(emailHandler.VerifyEmail))
	http.HandleFunc("/verified", users.AuthMiddleware(emailHandler.VerifiedUser))
	http.HandleFunc("/profile", users.AuthMiddleware(userHandler.MeHandler))
	http.HandleFunc("/addtoken", users.AuthMiddleware(tradeTokenHandler.AddToken))
	http.HandleFunc("/checkTradeToken", users.AuthMiddleware(tradeTokenHandler.CheckTradeToken))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
