package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	//_ "github.com/mattn/go-sqlite3"
)

var DATABASE = "/tmp/minitwit.db"
var PER_PAGE = 30
var DEBUG = true
var SECRET_KEY = "development key"

var (
	db    *sql.DB
	user  *string
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY"))) // TODO
)

type Student struct {
	Name string
}

// Returns a new connection to the database.
func connect_db() *sql.DB {
	db_, err := sql.Open("sqlite3", DATABASE)
	if err != nil {
		log.Fatal(err)
	}
	return db_
}

// Convenience method to return the db
func get_db() *sql.DB {
	return db
}

// Convenience method to look up the id for a username.
func get_user_id(username string) int {
	return 1 // TODO
}

// Convenience method to look up the the user.
func get_user() string {
	return *user
}

func followUser(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//username := vars["username"]
	//user := "jonas"
	whom_id := 1

	if whom_id <= 0 {
		w.WriteHeader(http.StatusNotFound)
	}
}

func before_req(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		db = connect_db()
		user = nil
		defer db.Close()
		handler(w, r)
	}
}

func public_timeline(w http.ResponseWriter, r *http.Request) {
	student := Student{
		Name: "Jo",
	}
	parsedTemp, _ := template.ParseFiles("test.html")
	err := parsedTemp.Execute(w, student)
	if err != nil {
		log.Println("Error executing template: ", err)
		return
	}
}

// Removes the current user as follower of the given user.
func unfollow_user(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	user := get_user()

	if len(user) <= 0 {
		w.WriteHeader(http.StatusUnauthorized)
	}

	whom_id := get_user_id(username)

	if whom_id <= 0 { // TODO how to check if i found nothing
		w.WriteHeader(http.StatusNotFound)
	}

	db := get_db()
	db.Exec("delete from follower where who_id=? and whom_id=?", get_user_id(user), whom_id)
	successMessage := []byte(fmt.Sprintf("You are no longer following %s", username))
	w.Write(successMessage)
	http.Redirect(w, r, fmt.Sprintf("/%v", vars["username"]), 302)
}

// Registers a new message for the user.
func add_message(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	user_id := session.Values["user_id"]

	if user_id == nil {
		w.WriteHeader(http.StatusUnauthorized)
	}

	message := r.FormValue("text")
	if len(message) > 0 {
		db := get_db()
		db.Exec("insert into message (author_id, text, pub_date, flagged) values (?, ?, ?, 0)", user_id, message, time.Now())
	}
	successMessage := []byte("Your message was recorded")
	w.Write(successMessage)

	http.Redirect(w, r, "timeline", 302)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/{username}/follow", before_req(followUser)).Methods("GET")
	r.HandleFunc("/", public_timeline).Methods("GET")

	r.HandleFunc("/{username}/unfollow", before_req(unfollow_user))
	r.HandleFunc("/add_message", before_req(add_message)).Methods("POST")

	http.ListenAndServe(":8080", r) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
