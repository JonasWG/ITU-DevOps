package main

import (
	"C"
	"database/sql"
	"fmt"
	"github.com/JonasWG/ITU-DevOps/structs"
	_ "github.com/JonasWG/ITU-DevOps/structs"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const DRIVER = "sqlite3"
const DATABASE = "./minitwit.db"
const PER_PAGE = 30
const DEBUG = true
const SECRET_KEY = "development key"

var templates map[string]*template.Template

func LoadTemplates() {
	var layoutTemplate = "templates/layout.gohtml"
	templates = make(map[string]*template.Template)

	templates["login"] = template.Must(template.ParseFiles(layoutTemplate, "templates/login.gohtml"))
	templates["signup"] = template.Must(template.ParseFiles(layoutTemplate, "templates/signup.gohtml"))
	templates["personal_timeline"] = template.Must(template.ParseFiles(layoutTemplate, "templates/personal_timeline.gohtml"))
	templates["public_timeline"] = template.Must(template.ParseFiles(layoutTemplate, "templates/public_timeline.gohtml"))
}

func GetUserByUsername(username string, db *sql.DB) (structs.User, error) {
	user := structs.User{}
	err := db.QueryRow("SELECT * FROM user WHERE username= ?", username, 1).
		Scan(&user.User_id, &user.Username, &user.Email, &user.Pw_hash)

	return user, err
}

func GetUserById(id int, db *sql.DB) (structs.User, error) {
	user := structs.User{}
	err := db.QueryRow("SELECT * FROM user WHERE user_id= ?", id, 1).
		Scan(&user.User_id, &user.Username, &user.Email, &user.Pw_hash)

	return user, err
}

func LoginHandler(store *sessions.CookieStore, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session_cookie")
		if r.Method == "GET" {
			if err := templates["login"].Execute(w, nil); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else if r.Method == "POST" {
			//TODO log in user
		} else {
			fmt.Println("could not match method:", r.Method)
		}
		userId := session.Values["user_id"]
		if isLoggedIn := userId != nil; isLoggedIn {
			http.Redirect(w, r, "/", http.StatusFound)
		}

		var errorMsg string
		if r.Method == "POST" {
			user, err := GetUserByUsername(r.FormValue("username"), db)
			if err != nil {
				errorMsg = "Invalid username"
			} else if err = bcrypt.CompareHashAndPassword([]byte(user.Pw_hash), []byte(r.FormValue("password"))); err != nil {
				errorMsg = "Invalid password"
			} else {
				session.AddFlash("You were logged in")
				session.Values["user_id"] = user.User_id
				err = session.Save(r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/", http.StatusFound)
			}
		}

		response := map[string]string{"error": errorMsg}
		log.Println(response)
		// TODO render login template with error
	})
}

func LogoutHandler(store *sessions.CookieStore, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session_cookie")
		session.Values["user_id"] = nil
		session.AddFlash("You were logged out")

		err := session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	})
}

func RegisterHandler(store *sessions.CookieStore, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session_cookie")

		if r.Method == "GET" {
			if err := templates["signup"].Execute(w, nil); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else if r.Method == "POST" {

		} else {
			fmt.Println("could not match method:", r.Method)
		}

		isLoggedIn := session.Values["user_id"] != nil
		if isLoggedIn {
			http.Redirect(w, r, "/", http.StatusFound)
		}

		var errorMsg string
		if r.Method == "POST" {
			if len(r.FormValue("username")) == 0 {
				errorMsg = "You have to enter a username"
			} else if len(r.FormValue("email")) == 0 || !strings.Contains(r.FormValue("email"), "@") {
				errorMsg = "You have to enter a valid email address"
			} else if len(r.FormValue("password")) == 0 {
				errorMsg = "You have to enter a password"
			} else if r.FormValue("password") != r.FormValue("password2") {
				errorMsg = "The passwords do not match"
			} else if user, _ := GetUserByUsername(r.FormValue("username"), db); user.Username == r.FormValue("username") {
				errorMsg = "This username is already taken"
			} else {
				statement, err := db.Prepare("INSERT INTO user (username, email, pw_hash) values (?,?,?)")
				if err != nil {
					log.Println(err)
					return
				}
				defer statement.Close()

				pass := r.FormValue("password")
				hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MinCost)
				if err != nil {
					log.Println(err)
					return
				}

				statement.Exec(r.FormValue("username"), r.FormValue("email"), hash)
				// TODO return successful registration status
			}
		}

		response := map[string]string{"error": errorMsg}
		log.Println(response)
		// w.Write([]byte(json.Marshal(response)))
		// TODO render register template with error
	})
}

func GetUserByIdHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, err := GetUserById(id, db)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		file := path.Join(".", "templates", "greet.html")
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			log.Fatal(err)
		}

		tmpl.Execute(w, user)
	})
}

func TestHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT COUNT(*) FROM user")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				log.Fatal(err)
			}
			fmt.Println(count)
		}
	})
}

func GetPostQuery(db *sql.DB) []structs.Post {
	rows, err := db.Query("select message.*, user.* from message, user where message.flagged = 0 and message.author_id = user.user_id order by message.pub_date desc limit 20")
	checkErr(err)

	var posts []structs.Post

	var message_id int
	var author_id int
	var text string
	var pub_date string
	var flagged int
	var user_id int
	var username string
	var email string
	var pw_hash string

	for rows.Next() {
		err = rows.Scan(&message_id, &author_id, &text, &pub_date, &flagged, &user_id, &username, &email, &pw_hash)
		checkErr(err)
		fmt.Println(message_id)
		fmt.Println(text)
		post := structs.Post{
			Message_id: message_id,
			Author_id:  author_id,
			Text:       text,
			Pub_date:   getTimeFromTimestamp(pub_date),
			Flagged:    flagged,
			Username:   username}
		posts = append(posts, post)
	}
	return posts
}

func getTimeFromTimestamp(timestamp string) string {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	checkErr(err)
	tm := time.Unix(i, 0)
	return tm.String()
}

func HomeHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var posts = GetPostQuery(db)
		content := structs.Content{Posts: posts, SignedIn: false}

		if err := templates["public_timeline"].Execute(w, content); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func BeforeRequestMiddleware(store *sessions.CookieStore, db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		mdfn := func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "session_cookie")
			userId := session.Values["user_id"]
			if userId != nil {
				id := userId.(int)
				user, err := GetUserById(id, db)
				if err != nil {
					log.Print(err)
				}

				session.Values["user_id"] = user.User_id
				err = session.Save(r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(mdfn)
	}
}

func initDb(driver string, datasource string) (*sql.DB, error) {
	db, err := sql.Open(driver, datasource)
	if err != nil {
		log.Fatal(err)
	}

	return db, db.Ping()
}

func main() {
	LoadTemplates()
	db, err := initDb(DRIVER, DATABASE)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// using the function
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(mydir)

	path, err := os.Executable()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fmt.Println(f.Name())
	}

	store := sessions.NewCookieStore([]byte(SECRET_KEY))

	r := mux.NewRouter()

	//CSS
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public/"))))
	r.Use(BeforeRequestMiddleware(store, db))

	r.Handle("/", HomeHandler(db)).Methods("GET")
	r.Handle("/public", TestHandler(db)).Methods("GET")
	r.Handle("/login", LoginHandler(store, db)).Methods("GET", "POST")
	r.Handle("/register", RegisterHandler(store, db)).Methods("GET", "POST")
	r.Handle("/logout", LogoutHandler(store, db)).Methods("GET")
	r.Handle("/add_message", TestHandler(db)).Methods("POST")
	r.Handle("{username}/", TestHandler(db)).Methods("GET")
	r.Handle("{username}/follow", TestHandler(db)).Methods("GET")
	r.Handle("{username}/unfollow", TestHandler(db)).Methods("GET")
	r.Handle("/test", TestHandler(db)).Methods("GET")
	r.Handle("/user/{id}", GetUserByIdHandler(db)).Methods("GET")

	//launch
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Printf("Open http://localhost:%s in the browser", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
