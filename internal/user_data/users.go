package user_data

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

type User struct {
	ID         int
	FirstName  string
	SecondName string
	Surname    string
	BirthDate  string
	Email      string
	Login      string
	Password   string
	Is_aproved bool
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
