package inmemory

import (
	"errors"

	"github.com/imthaghost/goland/zkp/internal/store"

	"github.com/k0kubun/pp/v3"
)

// InMemory is an inmemory database
//TODO should be switched over to a concurrent map
// or add concurrency safety
type InMemory struct {
	DB map[string]*store.User
}

// CreateUser will create a user in the in memory database
func (im *InMemory) CreateUser(u *store.User) error {
	im.DB[u.Username] = u

	// for pretty purposes :)
	pp.Print(im.DB)

	return nil
}

// GetUserByUsername will return the user by the given username
func (im *InMemory) GetUserByUsername(username string) (*store.User, error) {
	if val, ok := im.DB[username]; ok {
		return val, nil
	}

	return nil, errors.New("could not retrieve user")

}

// New will create a new interface to interface with an inmeory database.
func New() store.Service {
	return &InMemory{
		DB: make(map[string]*store.User),
	}
}
