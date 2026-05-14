package users

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mini-broker/internal/db"
	"net/http"
	"strings"

	/*"encoding/json"*/
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	/*"net/http"*/
	"time"
)

type Handler struct {
	DB *pgxpool.Pool
}

var jwtKey = []byte("session_key")

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
	reqPassword := hashing(req.Password)
	if getError != nil {
		log.Println(getError)
	}
	if user.Password != reqPassword {
		fmt.Println("Ошибка сверки данных", user.Password, " <> ", reqPassword)
		http.Error(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	token, tokenErr := GenerateToken(user.ID)
	if tokenErr != nil {
		log.Println(tokenErr)
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	} else {
		log.Println("Был сгенерирован токен при входе в систему. Токен = ", token)
	}

	// отправка токена клиента
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
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

func GenerateToken(userId int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseToken(tokenStr string) (int, error) {
	token, tokenErr := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if tokenErr != nil || !token.Valid {
		return 0, tokenErr
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	return userID, nil
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "AuthMiddleware: Missing token", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "token", tokenStr)

		next.ServeHTTP(w, r.WithContext(ctx))

	}
}

func (h *Handler) MeHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("user_id").(int)

	var username, firstName, secondName, surname, birthDate, email string
	err := h.DB.QueryRow(
		r.Context(),
		"SELECT first_name, second_name, surname, birth_date, email FROM users WHERE id=$1",
		userID,
	).Scan(&firstName, &secondName, &surname, &birthDate, &email)

	if err != nil {
		http.Error(w, "user not found", 404)
		return
	}

	username = firstName + " " + secondName + " " + surname
	json.NewEncoder(w).Encode(map[string]string{
		"Username":  username,
		"BirthDate": birthDate,
		"Email":     email,
	})
}
