package Token

import (
	"context"
	"encoding/json"
	"log"
	"mini-broker/internal/db"
	"mini-broker/internal/tbank"
	"mini-broker/internal/user_data"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Token struct {
	DB *pgxpool.Pool
}

type trade_token struct {
	Token string `json:"token"`
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := user_data.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", userID)

		next.ServeHTTP(w, r.WithContext(ctx))

	}
}

func (t *Token) AddToken(w http.ResponseWriter, r *http.Request) {
	var is_aproved bool
	var verify_token bool
	pool, _ := db.ConnectDB()
	sender := &user_data.SMTPEmailSender{}
	emailhandler := &user_data.EmailService{
		Sender: sender,
		DB:     pool,
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		log.Println("Пустой токен!")
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	user_id, err := user_data.ParseToken(token)

	log.Println("AddToken: Начинаем процесс добавления торгового токена!")

	var gettoken trade_token
	reqErr := json.NewDecoder(r.Body).Decode(&gettoken)
	if reqErr != nil {
		log.Println(reqErr)
	}

	trade_token_var := gettoken.Token
	is_aproved = emailhandler.VerifyCheck(user_id)
	log.Println("VerifyCheck: ", is_aproved)
	if is_aproved == false {
		log.Println("AddToken: Ошибка добавления токена! Email пользователя не верифицирован!")
		http.Error(w, "The user has not been verified", 403)
		return
	}
	verify_token = tbank.ConnectTbankWithTokenCheck(trade_token_var)
	log.Println("ConnectTbankWithTokenCheck: ", verify_token)
	if verify_token == false {
		log.Println("AddToken: Ошибка добавления токена! Токен", trade_token_var, "не прошел проверку")
		http.Error(w, "The token has not been verified", 401)
		return
	}
	log.Println("AddToken: Добавляем торговый токен для пользователя: ", user_id)

	conn, err := t.DB.Begin(context.Background())
	if err != nil {
		http.Error(w, "connect to DB error", 500)
		log.Println("connect to DB error", err.Error())
		return
	}
	defer conn.Rollback(context.Background())

	_, err = conn.Exec(context.Background(), `
		insert into user_tradetoken (user_id, token, create_time, status)
		values ($1, $2, now(), 'open')
		`, user_id, trade_token_var)
	if err != nil {
		http.Error(w, "insert token error", 500)
		log.Println("insert token error ", err.Error())
		return
	}
	err = conn.Commit(context.Background())
	if err != nil {
		http.Error(w, "commit insert token error", 500)
		log.Println("commit insert token error", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}
