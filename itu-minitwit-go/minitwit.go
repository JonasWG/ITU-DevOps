package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/{username}/unfollow", unfollow_user)
	r.HandleFunc("/add_message", add_message).Methods("POST")

	http.ListenAndServe(":80", r)
}

func unfollow_user(w http.ResponseWriter, r *http.Request) {

}

func add_message(w http.ResponseWriter, r *http.Request) {

}
