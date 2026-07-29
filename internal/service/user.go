package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"vortexuipro/internal/database"
	"vortexuipro/internal/domain"
	"vortexuipro/internal/events"
)

// UserService manages panel users and their proxy clients via database.
type UserService struct {
	eventBus events.Publisher
}

// NewUserService creates a new user service.
func NewUserService(bus events.Publisher) *UserService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &UserService{eventBus: bus}
}

// CreateUser adds a new user to the database.
func (s *UserService) CreateUser(username, email string, dataLimit int64, expiryTime int64) (*database.User, error) {
	user := &database.User{
		Username:   username,
		Email:      email,
		Status:     string(domain.UserActive),
		DataLimit:  dataLimit,
		ExpiryTime: expiryTime,
	}
	if err := database.CreateUser(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.eventBus.Publish(events.Event{
		Type:     events.UserCreated,
		UserID:   fmt.Sprintf("%d", user.ID),
		Username: user.Username,
	})
	return user, nil
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(id int64) (*database.User, error) {
	return database.GetUserByID(id)
}

// ListUsers returns all users with optional filtering.
func (s *UserService) ListUsers(adminID int64) ([]database.User, error) {
	return database.ListUsers(adminID)
}

// UpdateUser modifies an existing user.
func (s *UserService) UpdateUser(user *database.User) error {
	return database.UpdateUser(user)
}

// DeleteUser removes a user and related clients.
func (s *UserService) DeleteUser(id int64) error {
	u, err := database.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if err := database.DeleteUser(id); err != nil {
		return err
	}
	s.eventBus.Publish(events.Event{
		Type:     events.UserDeleted,
		UserID:   fmt.Sprintf("%d", id),
		Username: u.Username,
	})
	return nil
}

// AddClient adds a proxy client for a user.
func (s *UserService) AddClient(userID int64, inboundID int64, email string) (*database.Client, error) {
	client := &database.Client{
		ID:         uuid.New().String(),
		UserID:     userID,
		InboundID:  inboundID,
		Email:      email,
		Enable:     true,
	}
	if err := database.CreateClient(client); err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	return client, nil
}

// GetClient retrieves a client by ID.
func (s *UserService) GetClient(id string) (*database.Client, error) {
	return database.GetClientByID(id)
}

// ListClients returns all clients for a user or inbound.
func (s *UserService) ListClients(userID, inboundID int64) ([]database.Client, error) {
	return database.ListClients(userID, inboundID)
}

// DeleteClient removes a client.
func (s *UserService) DeleteClient(id string) error {
	return database.DeleteClient(id)
}

// CheckExpirations checks and updates expired users.
func (s *UserService) CheckExpirations() {
	users, err := database.ListUsers(0)
	if err != nil {
		return
	}
	now := time.Now().UnixMilli()
	for _, u := range users {
		if u.ExpiryTime > 0 && u.ExpiryTime <= now && u.Status == string(domain.UserActive) {
			u.Status = string(domain.UserExpired)
			database.UpdateUser(&u)
			s.eventBus.Publish(events.Event{
				Type:     events.UserExpired,
				UserID:   fmt.Sprintf("%d", u.ID),
				Username: u.Username,
			})
		}
	}
}

// RecordTraffic updates traffic counters for a user.
func (s *UserService) RecordTraffic(email string, up, down int64) {
	client, err := database.GetClientByEmail(email)
	if err != nil {
		return
	}
	user, err := database.GetUserByID(client.UserID)
	if err != nil {
		return
	}
	user.TrafficUp += up
	user.TrafficDown += down
	database.UpdateUser(user)

	// Check data limit
	if user.DataLimit > 0 && (user.TrafficUp+user.TrafficDown) >= user.DataLimit {
		user.Status = string(domain.UserLimited)
		database.UpdateUser(user)
		s.eventBus.Publish(events.Event{
			Type:   events.UserLimited,
			UserID: fmt.Sprintf("%d", user.ID),
		})
	}
}
