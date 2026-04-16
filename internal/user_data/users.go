package user_data

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mini-broker/internal/db"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID         int    `json:"id"`
	FirstName  string `json:"firstname"`
	SecondName string `json:"secondname"`
	Surname    string `json:"surname"`
	BirthDate  string `json:"birthdate"`
	Email      string `json:"email"`
	Login      string `json:"login"`
	Password   string `json:"password"`
	Is_aproved bool   `json:"is_aproved"`
}

func GetUsers(db *pgxpool.Pool) ([]User, error) {
	rows, err := db.Query(context.Background(), `
		SELECT *
		FROM users
	`)
	if err != nil {
		return nil, fmt.Errorf("Ошибка получения данных о пользователях: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.SecondName,
			&user.Surname,
			&user.BirthDate,
			&user.Email,
			&user.Login,
			&user.Password,
			&user.Is_aproved,
		)
		if err != nil {
			return nil, fmt.Errorf("Ошибка сканирования пользователей: %w", err)
		}
		users = append(users, user)
	}
	log.Println("Найдено пользователей:", len(users))

	return users, nil
}

func AddInputUser(db *pgxpool.Pool) error {
	query := `
		INSERT INTO users (first_name, second_name, surname, birth_date, email, login, password, is_aproved)	
		values ($1, $2, $3, $4, $5, $6, $7, false)
		returning id;
	`
	var id int
	user, err := InputUser()
	if err != nil {
		log.Println(err)
	}
	user.Password = hashing(user.Password)
	error := db.QueryRow(context.Background(), query, user.FirstName, user.SecondName, user.Surname, user.BirthDate, user.Email, user.Login, user.Password).Scan(&id)
	if error != nil {
		log.Println(error)
	} else {
		log.Println("Добавлен новый пользователь! ID: ", id)
	}
	return nil
}

func AddUserByForm(db *pgxpool.Pool, user User) error {
	query := `
		INSERT INTO users (first_name, second_name, surname, birth_date, email, login, password, is_aproved)	
		values ($1, $2, $3, $4, $5, $6, $7, false)
		returning id;
	`
	var id int

	user.Password = hashing(user.Password)
	error := db.QueryRow(context.Background(), query, user.FirstName, user.SecondName, user.Surname, user.BirthDate, user.Email, user.Login, user.Password).Scan(&id)
	if error != nil {
		log.Println(error)
	} else {
		log.Println("Добавлен новый пользователь! ID: ", id)
	}
	return nil
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pool, err := db.ConnectDB()
	if err != nil {
		log.Println(err)
	}

	var req User
	reqErr := json.NewDecoder(r.Body).Decode(&req)
	if reqErr != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		fmt.Println("Invalid request", reqErr)
		return
	}

	if req.FirstName == "" || req.Surname == "" || req.BirthDate == "" ||
		req.Email == "" || req.Login == "" || req.Password == "" {
		http.Error(w, "Empty fields", http.StatusBadRequest)
		fmt.Println("Empty fields", reqErr)
		return
	}

	addUserErr := AddUserByForm(pool, req)
	if addUserErr != nil {
		log.Println(addUserErr)
	}

}

func InputUser() (user User, err error) {
	fmt.Println("Ввод данных пользователя")
	fmt.Println("Имя:")
	fmt.Scan(&user.FirstName)
	fmt.Println("Отчество:")
	fmt.Scan(&user.SecondName)
	fmt.Println("Фамилия:")
	fmt.Scan(&user.Surname)
	fmt.Println("Дата рождения:")
	fmt.Scan(&user.BirthDate)
	fmt.Println("Email")
	fmt.Scan(&user.Email)
	fmt.Println("Login")
	fmt.Scan(&user.Login)
	fmt.Println("Password")
	fmt.Scan(&user.Password)

	return user, nil
}
