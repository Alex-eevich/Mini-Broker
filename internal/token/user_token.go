package token

import (
	"context"
	"encoding/json"
	"log"
	"mini-broker/internal/db"
	"mini-broker/internal/tbank"
	"mini-broker/internal/users"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
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

		userID, err := users.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", userID)

		next.ServeHTTP(w, r.WithContext(ctx))

	}
}

func (t *Token) AddToken(w http.ResponseWriter, r *http.Request) {
	var isAproved bool
	var verifyToken bool
	pool, _ := db.ConnectDB()
	sender := &users.SMTPEmailSender{}
	emailHandler := &users.EmailService{
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
	userId, userIdErr := users.ParseToken(token)
	if userIdErr != nil {
		log.Println(userIdErr)
	}

	log.Println("AddToken: Начинаем процесс добавления торгового токена!")

	var getToken trade_token
	reqErr := json.NewDecoder(r.Body).Decode(&getToken)
	if reqErr != nil {
		log.Println(reqErr)
	}

	tradeTokenVar := getToken.Token
	isAproved = emailHandler.VerifyCheck(userId)
	log.Println("VerifyCheck: ", isAproved)
	if isAproved == false {
		log.Println("AddToken: Ошибка добавления токена! Email пользователя не верифицирован!")
		http.Error(w, "The user has not been verified", 403)
		return
	}
	verifyToken = tbank.ConnectTbankWithTokenCheck(tradeTokenVar)
	log.Println("ConnectTbankWithTokenCheck: ", verifyToken)
	if verifyToken == false {
		log.Println("AddToken: Ошибка добавления токена! Токен", tradeTokenVar, "не прошел проверку")
		http.Error(w, "The token has not been verified", 401)
		return
	}
	log.Println("AddToken: Добавляем торговый токен для пользователя: ", userId)

	conn, connErr := t.DB.Begin(context.Background())
	if connErr != nil {
		http.Error(w, "connect to DB error", 500)
		log.Println("connect to DB error", connErr.Error())
		return
	}
	defer conn.Rollback(context.Background())

	checkToken := CheckToken(userId, conn)
	if checkToken == false {
		log.Println("Токен уже есть у пользователя! ")
		http.Error(w, "У вас уже есть токен!", 402)
		return
	}

	_, insertErr := conn.Exec(context.Background(), `
		insert into user_tradetoken (user_id, token, create_time, status)
		values ($1, $2, now(), 'open')
		`, userId, tradeTokenVar)
	if insertErr != nil {
		http.Error(w, "insert token error", 500)
		log.Println("insert token error ", insertErr.Error())
		return
	}
	commitErr := conn.Commit(context.Background())
	if commitErr != nil {
		http.Error(w, "commit insert token error", 500)
		log.Println("commit insert token error", commitErr.Error())
		return
	}

	client := tbank.NewClientFromToken(tradeTokenVar)
	dbConn := tbank.DB_connect{
		DB: pool,
	}
	accounts, accountsErr := client.GetListAccounts()
	if accountsErr != nil {
		http.Error(w, "get account error", 500)
		return
	}
	if accounts == "" {
		addedAccount, err := client.OpenSandboxAccount()
		if err != nil {
			http.Error(w, "add account error", 500)
			return
		}
		log.Println("AddToken.AddAccount: Добавлен торговый счет для пользователя", userId)
		addRes := dbConn.AddTradeAccountDB(addedAccount, userId)
		if addRes != nil {
			log.Println("AddTradeAccountDB", addRes)
			http.Error(w, "AddTradeAccountDB: ошибка добавления торгового счета в БД", 500)
			return
		} else {
			log.Println("AddTradeAccountDB:", addedAccount, " добавлен в БД")
		}
	} else {
		addRes := dbConn.AddTradeAccountDB(accounts, userId)
		if addRes != nil {
			log.Println("AddTradeAccountDB", addRes)
			http.Error(w, "AddTradeAccountDB: ошибка добавления торгового счета в БД", 500)
			return
		} else {
			log.Println("AddTradeAccountDB:", accounts, " добавлен в БД")
		}
	}
	w.WriteHeader(http.StatusOK)
}

func CheckToken(userId int, conn pgx.Tx) bool {
	var tokenBody string
	query := `
		select id from user_tradetoken
		where user_id = $1;`
	selectErr := conn.QueryRow(context.Background(), query, userId).Scan(&tokenBody)
	if selectErr != nil {
		log.Println(selectErr)
	}
	if tokenBody != "" {
		return false
	} else {
		return true
	}
}

func (t *Token) CheckTradeToken(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user_id").(int)
	pool, _ := db.ConnectDB()
	defer pool.Close()
	conn, connErr := t.DB.Begin(context.Background())
	if connErr != nil {
		http.Error(w, "connect to DB error", 500)
		log.Println("connect to DB error", connErr.Error())
		return
	}
	defer conn.Rollback(context.Background())
	check := CheckToken(userId, conn)
	if check == true {
		http.Error(w, "CheckToken error", 403)
		return
	}
	w.WriteHeader(http.StatusOK)
}
