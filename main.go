package main

import (
	//database/sql does not know what flavor of sql you intend to use. So we also import the go-sql-driver/mysql driver to tell it we're using mysql instead of postgres or mssql or something else.
	"database/sql"
	"encoding/json"
	"fmt"
	// Because we don't call mysql directly, but database/sql uses it, we need to import it and ignore it with the underscore in front
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

//the port that the api is running on
var port = "8020"

//a global database variable as a pointer. We use a pointer so we don't have to create a bunch of copies of our database connection.
var db *sql.DB

//A custom struct to hold the fields from the database
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

func main() {

	//NEVER COMMIT REAL CREDENTIALS TO GITHUB.
	database, err := sql.Open("mysql", "root:passwd@tcp(127.0.0.1:3306)/example-db")
	if err != nil {
		panic(err)
	}
	//assign the database connection we created to the global variable db
	db = database

	//Always close your connection with a defer right after opening. Defer will run the db.Close when the function main() completes.
	defer db.Close()

	//define which function handles which route
	http.HandleFunc("/", handleRequest)

	//need an application to run to constantly listen
	fmt.Printf("Listening on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

//This is the handler for the root path "/"
func handleRequest(w http.ResponseWriter, r *http.Request) {
	//if the HTTP method is a GET request, we want to get the users from the DB
	if r.Method == http.MethodGet {
		//define a slice of users to hold all the rows that come back from the database.
		users := []User{}

		//write a query to send to the database. It's best practice to list all columns individually instead of `SELECT *` because the columns in the table can change.
		query := `SELECT id, first_name, last_name, email FROM user`

		//Use db.Query when you expect to get back multiple rows
		rows, err := db.Query(query)
		//always, always check your errors
		if err != nil {
			fmt.Println(err)
			//this return statement prevents the rest of the code below here from executing.
			return
		}

		// This structure is looping over the rows returned from the db.Query using rows.Next(). rows.Next() returns true if there's still a row left to deal with
		for rows.Next() {
			//create a user to store the row into.
			var user User
			//scan each column
			err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)
			//always check your errors
			if err != nil {
				fmt.Println(err)
				return
			}
			//add the user we created from this row to the slice of users we created earlier
			users = append(users, user)
		}

		//add an http status header reflecting the outcome of the request to the ResponseWriter
		w.WriteHeader(http.StatusOK)
		//Make the ResponseWriter into a json encoder, then encode the users slice into json and send the response.
		json.NewEncoder(w).Encode(users)

	}

	//If the request was an http POST request
	if r.Method == http.MethodPost {
		//create a user to hold the incoming request
		var user User
		//read the request body with a json decoder and store into the user just created
		json.NewDecoder(r.Body).Decode(&user)

		//insert query with `?` to parameterize values to protect from sql injection
		query := `INSERT INTO user (first_name, last_name, email) values (?,?,?)`
		//using db.Exec for inserts returns a Result, not a row. Give it a query plus the parameters that will replace each `?` in the query string
		res, err := db.Exec(query, user.FirstName, user.LastName, user.Email)
		//for real though, catch those errors.
		if err != nil {
			fmt.Println(err)
			return
		}
		//if there wasn't an error, then there was no problem connecting to the database and running the query.
		// You can then use the Result, res, to find out what happened. LastInsertId returns the auto-incremented id for the item that you just saved.
		id, err := res.LastInsertId()
		//Catchin' those errors
		if err != nil {
			fmt.Println(err)
			return
		}
		// now that you have the last inserted ID, you can save it to the user that came in on the original request.
		user.ID = id

		//Because we successfully saved the user, let the caller know the item was created with a HTTP status code 201 Created
		w.WriteHeader(http.StatusCreated)

		//Use the writer (w) to create a json encoder and encode the user that was saved and respond to the caller with json.
		json.NewEncoder(w).Encode(user)
	}
}
