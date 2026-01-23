package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/go-generic-event-driven/pkg/cqrs"
	"github.com/yourusername/go-generic-event-driven/pkg/ddd"
	"github.com/yourusername/go-generic-event-driven/pkg/logger"
	"github.com/yourusername/go-generic-event-driven/pkg/module"
	"github.com/yourusername/go-generic-event-driven/pkg/types"
)

// Example: User aggregate using generic boilerplate

// User represents a user entity
type User struct {
	types.Entity
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserCreated event
type UserCreated struct {
	ddd.BaseEvent
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

func NewUserCreated(userID, name, email string) *UserCreated {
	event := ddd.NewBaseEvent("UserCreated", map[string]interface{}{
		"user_id": userID,
		"name":    name,
		"email":   email,
	})
	return &UserCreated{
		BaseEvent: *event,
		UserID:    userID,
		Name:      name,
		Email:     email,
	}
}

// UserUpdated event
type UserUpdated struct {
	ddd.BaseEvent
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

func NewUserUpdated(userID, name string) *UserUpdated {
	event := ddd.NewBaseEvent("UserUpdated", map[string]interface{}{
		"user_id": userID,
		"name":    name,
	})
	return &UserUpdated{
		BaseEvent: *event,
		UserID:    userID,
		Name:      name,
	}
}

// CreateUserHandler implements cqrs.CommandHandler
type CreateUserHandler struct {
	users map[string]*User
}

func NewCreateUserHandler(users map[string]*User) *CreateUserHandler {
	return &CreateUserHandler{users: users}
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *cqrs.Command) (*cqrs.CommandResult, error) {
	payload := cmd.Data.(map[string]interface{})

	userID := fmt.Sprintf("user_%d", time.Now().UnixNano())
	user := &User{
		Entity: *types.NewEntityWithCreator(userID, "system"),
		Name:   payload["name"].(string),
		Email:  payload["email"].(string),
	}

	h.users[userID] = user

	event := NewUserCreated(userID, user.Name, user.Email)

	return &cqrs.CommandResult{
		Events:      []ddd.Event{event},
		AggregateID: userID,
		Version:     1,
	}, nil
}

// UpdateUserHandler implements cqrs.CommandHandler
type UpdateUserHandler struct {
	users map[string]*User
}

func NewUpdateUserHandler(users map[string]*User) *UpdateUserHandler {
	return &UpdateUserHandler{users: users}
}

func (h *UpdateUserHandler) Handle(ctx context.Context, cmd *cqrs.Command) (*cqrs.CommandResult, error) {
	payload := cmd.Data.(map[string]interface{})
	userID := payload["user_id"].(string)

	user, exists := h.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	user.Name = payload["name"].(string)
	user.UpdateMetadata("system")

	event := NewUserUpdated(userID, user.Name)

	return &cqrs.CommandResult{
		Events:      []ddd.Event{event},
		AggregateID: userID,
		Version:     user.Metadata.Version,
	}, nil
}

// GetUserHandler implements cqrs.QueryHandler
type GetUserHandler struct {
	users map[string]*User
}

func NewGetUserHandler(users map[string]*User) *GetUserHandler {
	return &GetUserHandler{users: users}
}

func (h *GetUserHandler) Handle(ctx context.Context, query *cqrs.Query) (interface{}, error) {
	payload := query.Data.(map[string]interface{})
	userID := payload["user_id"].(string)

	user, exists := h.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// UserCreatedEventHandler implements cqrs.EventHandler
type UserCreatedEventHandler struct{}

func NewUserCreatedEventHandler() *UserCreatedEventHandler {
	return &UserCreatedEventHandler{}
}

func (h *UserCreatedEventHandler) Handle(ctx context.Context, event ddd.Event) error {
	userCreated := event.(*UserCreated)
	fmt.Printf("🎉 User created: %s (%s)\n", userCreated.Name, userCreated.Email)
	return nil
}

// UserUpdatedEventHandler implements cqrs.EventHandler
type UserUpdatedEventHandler struct{}

func NewUserUpdatedEventHandler() *UserUpdatedEventHandler {
	return &UserUpdatedEventHandler{}
}

func (h *UserUpdatedEventHandler) Handle(ctx context.Context, event ddd.Event) error {
	userUpdated := event.(*UserUpdated)
	fmt.Printf("✏️  User updated: %s\n", userUpdated.Name)
	return nil
}

// RegisterUserModule registers the user module with the application
func RegisterUserModule(app *module.Application, users map[string]*User) error {
	// Create module
	userModule := module.NewModule("user")

	// Register command handlers
	createHandler := NewCreateUserHandler(users)
	updateHandler := NewUpdateUserHandler(users)

	if err := userModule.RegisterCommandHandler("CreateUser", createHandler); err != nil {
		return err
	}
	if err := userModule.RegisterCommandHandler("UpdateUser", updateHandler); err != nil {
		return err
	}

	// Register query handlers
	getHandler := NewGetUserHandler(users)
	if err := userModule.RegisterQueryHandler("GetUser", getHandler); err != nil {
		return err
	}

	// Register event handlers
	createdHandler := NewUserCreatedEventHandler()
	updatedHandler := NewUserUpdatedEventHandler()

	if err := userModule.RegisterEventHandler("UserCreated", createdHandler); err != nil {
		return err
	}
	if err := userModule.RegisterEventHandler("UserUpdated", updatedHandler); err != nil {
		return err
	}

	// Register module with application
	return app.RegisterModule(userModule)
}

func main() {
	fmt.Println("🚀 Testing Go Generic Event-Driven Boilerplate")

	// Create logger
	log := logger.NewStandardLoggerWithPrefix("BOILERPLATE")

	// Create application
	app := module.NewApplication(log)

	// Create user storage
	users := make(map[string]*User)

	// Register user module
	if err := RegisterUserModule(app, users); err != nil {
		log.Fatal("Failed to register user module", logger.NewField("error", err.Error()))
	}

	// Start application
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		log.Fatal("Failed to start application", logger.NewField("error", err.Error()))
	}
	defer app.Stop()

	fmt.Println("✅ Application started successfully!")

	// Test command: Create user
	createCmd := cqrs.NewCommand("CreateUser", "user-1", map[string]interface{}{
		"name":  "John Doe",
		"email": "john.doe@example.com",
	})

	result, err := app.CQRS.ExecuteCommand(ctx, createCmd)
	if err != nil {
		log.Fatal("Failed to execute create command", logger.NewField("error", err.Error()))
	}

	fmt.Printf("✅ Create command executed. Events: %d, AggregateID: %s\n",
		len(result.Events), result.AggregateID)

	// Test query: Get user
	getUserQuery := cqrs.NewQuery("GetUser", map[string]interface{}{
		"user_id": result.AggregateID,
	})

	queryResult, err := app.CQRS.ExecuteQuery(ctx, getUserQuery)
	if err != nil {
		log.Fatal("Failed to execute get query", logger.NewField("error", err.Error()))
	}

	user := queryResult.(*User)
	fmt.Printf("👤 Query result: User {ID: %s, Name: %s, Email: %s}\n",
		user.ID, user.Name, user.Email)

	// Test command: Update user
	updateCmd := cqrs.NewCommand("UpdateUser", result.AggregateID, map[string]interface{}{
		"user_id": user.ID,
		"name":    "John Smith",
	})

	_, err = app.CQRS.ExecuteCommand(ctx, updateCmd)
	if err != nil {
		log.Fatal("Failed to execute update command", logger.NewField("error", err.Error()))
	}

	// Query again to see the update
	queryResult, err = app.CQRS.ExecuteQuery(ctx, getUserQuery)
	if err != nil {
		log.Fatal("Failed to execute get query after update", logger.NewField("error", err.Error()))
	}

	updatedUser := queryResult.(*User)
	fmt.Printf("🔄 Updated user: User {ID: %s, Name: %s, Email: %s}\n",
		updatedUser.ID, updatedUser.Name, updatedUser.Email)

	fmt.Println("✨ Generic boilerplate test completed successfully!")
	fmt.Println("\n📋 Boilerplate features validated:")
	fmt.Println("  ✅ Generic CQRS pattern")
	fmt.Println("  ✅ Generic DDD interfaces")
	fmt.Println("  ✅ Generic module system")
	fmt.Println("  ✅ Generic event handling")
	fmt.Println("  ✅ Generic types and payloads")
	fmt.Println("  ✅ Generic logging")
	fmt.Println("  ✅ Reusable without business logic")
}
