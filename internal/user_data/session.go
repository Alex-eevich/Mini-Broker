package user_data

import (
	"context"
	"encoding/json"
	"log"
	"mini-broker/internal/db"
	"net/http"

	/*"encoding/json"*/
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	/*"net/http"*/
	"time"
)

var jwtKey = []byte("session_key")

func GenerateToken(user_id int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user_id,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}

func GetUserByLogin(db *pgxpool.Pool, login string) (*User, error) {
	query := `
	Select id, first_name, second_name, surname, birth_date, email, login, "password", is_aproved from users u where login = $1;
	`
	var user User
	getError := db.QueryRow(context.Background(), query, login).Scan(&user.ID, &user.FirstName, &user.SecondName, &user.Surname, &user.BirthDate, &user.Email, &user.Login, &user.Password, &user.Is_aproved)
	if getError != nil {
		log.Println("Ошибка проверки наличия пользователя:", getError)
	} else {
		log.Println(user)
	}
	return &user, nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	pool, err := db.ConnectDB()
	if err != nil {
		log.Println(err)
	}
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	user, getError := GetUserByLogin(pool, req.Login)
	if getError != nil {
		log.Println(getError)
	}
	if user.Password != req.Password {
		http.Error(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	// отправка токена клиента
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
