package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

var port = "8020"
var db *sql.DB

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

var users = make([]User, 0)

func main() {

	database, err := sql.Open("mysql", "root:passwd@tcp(127.0.0.1:3306)/example-db")
	if err != nil {
		panic(err)
	}
	db = database
	defer db.Close()

	http.HandleFunc("/", handleRequest)

	//need an application to run to constantly listen
	fmt.Printf("Listening on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		users := []User{}

		query := `SELECT * FROM user`
		rows, err := db.Query(query)
		if err != nil {
			fmt.Println(err)
			return
		}

		for rows.Next() {
			var user User
			err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)
			if err != nil {
				fmt.Println(err)
				return
			}
			users = append(users, user)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)

	}

	if r.Method == http.MethodPost {
		var user User
		json.NewDecoder(r.Body).Decode(&user)

		query := `INSERT INTO user (first_name, last_name, email) values (?,?,?)`
		res, err := db.Exec(query, user.FirstName, user.LastName, user.Email)
		if err != nil {
			fmt.Println(err)
			return
		}
		id, err := res.LastInsertId()
		if err != nil {
			fmt.Println(err)
			return
		}
		user.ID = id

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}
