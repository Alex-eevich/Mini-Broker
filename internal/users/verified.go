package users

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"net/http"
	"net/smtp"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailSender interface {
	Send(email string, body string) error
}

type EmailService struct {
	Sender EmailSender
	DB     *pgxpool.Pool
}

func (s *EmailService) SendEmail(email, body string) {
	go func() {
		err := s.Sender.Send(email, body)
		if err != nil {
			log.Println("SendEmail (go func)", err)
		}
	}()
}

type SMTPEmailSender struct{}

func generateVerifyToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *SMTPEmailSender) Send(email string, body string) error {
	myEmail := "alexeevich.2323655@gmail.com"

	auth := smtp.PlainAuth(
		"",
		myEmail,
		"ndyp yyqv laze bjaf",
		"smtp.gmail.com",
	)

	msg := "Subject: Verify Email\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		body

	log.Println("Sending email to:", email)

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		myEmail,
		[]string{email},
		[]byte(msg),
	)
}

func (s *EmailService) VerifiedUser(w http.ResponseWriter, r *http.Request) {

	userId := r.Context().Value("userId").(int)
	tokenStr, genErr := generateVerifyToken()
	if genErr != nil {
		http.Error(w, "Error create verify token", http.StatusUnauthorized)
	}

	query := `
	Select email from users u where id = $1;
	`
	var email string
	var messageType string
	getErr := s.DB.QueryRow(context.Background(), query, userId).Scan(&email)
	if getErr != nil {
		log.Println("Ошибка проверки наличия пользователя:", getErr)
	}
	link := "http://localhost:8080/verify-email?token=" + tokenStr
	body := "Verify email: " + link
	sendErr := s.Sender.Send(
		email,
		body,
	)
	if sendErr != nil {
		log.Println("SendError", sendErr)
	} else {
		log.Println("Отправлено сообщение на Email:", email)
		messageType = "Mail_Verified"
		s.InsertMessage(strconv.Itoa(userId), tokenStr, messageType, email)
		w.WriteHeader(http.StatusOK)
	}
}

func hashing(body string) string {
	hashBytes := sha256.Sum256([]byte(body))
	return hex.EncodeToString(hashBytes[:])
}

func (s *EmailService) InsertMessage(userId, token, messageType, email string) {
	query := `
		INSERT INTO mail_message (type, user_id, body, email,create_time, status, usedate)	
		values ($1, $2, $3, $4, now(), 'send', NULL)
		returning id;
	`
	var id int
	error := s.DB.QueryRow(context.Background(), query, messageType, userId, token, email).Scan(&id)
	if error != nil {
		log.Println(error)
	} else {
		log.Println("Добавлено новое mail_message! ID: ", id)
	}
}

func (s *EmailService) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token").(string)

	var userID int
	query := `
		select user_id from mail_message
		where body = $1 and status = 'send';`
	err := s.DB.QueryRow(context.Background(), query, token).Scan(&userID)
	if err != nil {
		log.Println("Токен не дейстивтелен! ", err)
	}

	tx, txErr := s.DB.Begin(context.Background())
	if txErr != nil {
		http.Error(w, "tx error", 500)
		return
	}
	defer tx.Rollback(context.Background())
	_, err = tx.Exec(context.Background(), `
		update users set is_aproved = true
		where id = $1
		`, userID)
	if err != nil {
		log.Println("Не смогли верифицировать в БД: ", err)
		return
	}
	_, err = tx.Exec(context.Background(), `
		update mail_message 
		set status = 'used', usedate = now()
		where body = $1
		`, token)
	if err != nil {
		log.Println("Не смогли закрыть токен в БД: ", err)
		return
	}
	err = tx.Commit(context.Background())
	if err != nil {
		http.Error(w, "commit error", 500)
		return
	}

	html := `
	<!DOCTYPE html>
	<html>
	<head>
	    <title>Email verified</title>
		<meta http-equiv="refresh" content="3;url=/authorization.html">
	</head>
	<body>
	    <h2>Email successfully verified!</h2>
	    <p>You can close this page</p>
		<div id="VerifyStatus"></div>
	</body>
	</html>
	`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *EmailService) VerifyCheck(user_id int) bool {
	var isAproved bool
	query := `
		select is_aproved from users u where id = $1;`
	selectErr := s.DB.QueryRow(context.Background(), query, user_id).Scan(&isAproved)
	if selectErr != nil {
		log.Println("VerifyCheck: Ошибка проверки верификации пользователя", selectErr)
		return false
	}
	if isAproved == true {
		return true
	} else {
		return false
	}
}
