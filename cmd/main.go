package main

import (
	"log"
	"mini-broker/internal/user_data"
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./internal/web")))
	http.HandleFunc("/register", user_data.RegisterUser)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

	/*pool, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	users, getErr := user_data.GetUsers(pool)
	if getErr != nil {
		log.Fatal(getErr)
	}
	for user := range users {
		fmt.Println(users[user].ID, users[user].FirstName, users[user].SecondName, users[user].Surname, users[user].BirthDate)
	}
	addError := user_data.AddUser(pool)
	if addError != nil {
		log.Fatal("Ошибка добавления пользователя!", addError)
	}
	*/
}
