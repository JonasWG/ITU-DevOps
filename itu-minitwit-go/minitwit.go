package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/{username}/unfollow", unfollow_user)
	r.HandleFunc("/add_message", add_message).Methods("POST")

	http.ListenAndServe(":80", r)
}

// Convenience method to return the db
func get_db() (db *sql.DB) {
	return db
}

// Convenience method to look up the id for a username.
func get_user_id(username string) int {
	return 1 // TODO
}

// Convenience method to look up the the user.
func get_user(r *http.Request) string {
	return "alma" // TODO
}

// Removes the current user as follower of the given user.
func unfollow_user(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	user := get_user(r)

	if len(user) <= 0 {
		w.WriteHeader(http.StatusUnauthorized)
	}

	whom_id := get_user_id(username)

	if whom_id <= 0 { // TODO how to check if i found nothing
		w.WriteHeader(http.StatusNotFound)
	}

	db := get_db()
	db.Exec("delete from follower where who_id=? and whom_id=?", get_user_id(user), whom_id)
	msg := []byte(fmt.Sprintf("You are no longer following %s", username))
	w.Write(msg)
	http.Redirect(w, r, fmt.Sprintf("/%v", vars["username"]), 302)
}

// Registers a new message for the user.
func add_message(w http.ResponseWriter, r *http.Request) {

}
