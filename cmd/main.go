package main

import (
	"fmt"
	"log"
	"mini-broker/internal/db"
	"mini-broker/internal/user_data"
)

func main() {
	pool, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	users, err := user_data.GetUsers(pool)
	if err != nil {
		log.Fatal(err)
	}
	for user := range users {
		fmt.Println(users[user].ID, users[user].FirstName, users[user].SecondName, users[user].Surname, users[user].BirthDate)
	}
}
