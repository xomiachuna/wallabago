package core

type User struct {
	ID       string
	IsAdmin  bool
	Username string
}

type CreatedUser struct {
	Username string
	Password string
}
