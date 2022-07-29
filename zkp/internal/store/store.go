package store

// Service describes how we interface with the database
type Service interface {
	CreateUser(*User) error
	GetUserByUsername(username string) (*User, error)
}
