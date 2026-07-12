package userdb

import (
	"context"
	"sync"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/userrepo"
)

// MemoryUser is an in-memory implementation of userrepo.Repository.
type MemoryUser struct {
	mu     sync.RWMutex
	nextID uint
	store  map[uint]model.User
}

// ProvideMemory creates an empty in-memory user store.
func ProvideMemory() *MemoryUser {
	return &MemoryUser{
		mu:     sync.RWMutex{},
		nextID: 0,
		store:  make(map[uint]model.User),
	}
}

// Create persists a new user and returns it with its assigned ID.
func (m *MemoryUser) Create(_ context.Context, user model.User) (model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.store {
		if u.Username == user.Username {
			return model.User{}, userrepo.ErrDuplicateUsername
		}
	}

	m.nextID++
	user.ID = m.nextID
	m.store[user.ID] = user

	return user, nil
}

// FindByUsername returns the user with the given username.
func (m *MemoryUser) FindByUsername(_ context.Context, username string) (model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.store {
		if u.Username == username {
			return u, nil
		}
	}

	return model.User{}, userrepo.ErrUserNotFound
}

// FindByID returns the user with the given id.
func (m *MemoryUser) FindByID(_ context.Context, id uint) (model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if u, ok := m.store[id]; ok {
		return u, nil
	}

	return model.User{}, userrepo.ErrUserNotFound
}

// FindBySubject returns the user federated from provider with the given subject.
func (m *MemoryUser) FindBySubject(
	_ context.Context, provider model.Provider, subject string,
) (model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.store {
		if u.Provider == provider && u.Subject == subject {
			return u, nil
		}
	}

	return model.User{}, userrepo.ErrUserNotFound
}

// List returns all users.
func (m *MemoryUser) List(_ context.Context) ([]model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]model.User, 0, len(m.store))
	for _, u := range m.store {
		users = append(users, u)
	}

	return users, nil
}

// SetRole updates the role of the user with the given id.
func (m *MemoryUser) SetRole(_ context.Context, id uint, role model.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, ok := m.store[id]
	if !ok {
		return userrepo.ErrUserNotFound
	}

	u.Role = role
	m.store[id] = u

	return nil
}

// Delete removes the user with the given id.
func (m *MemoryUser) Delete(_ context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.store[id]; !ok {
		return userrepo.ErrUserNotFound
	}

	delete(m.store, id)

	return nil
}
