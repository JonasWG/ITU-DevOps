package structs

type Content struct {
	SignedIn bool
	Posts    []Post
}

type User struct {
	User_id  int
	Username string
	Email    string
	Pw_hash  string
}

type Post struct {
	Username   string
	Message_id int
	Author_id  int
	Text       string
	Pub_date   string
	Flagged    int
}
