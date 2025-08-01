package repository

import (
	"context"
	"fmt"
	"testing"

	"user-svc/config"
	models "user-svc/internal/domain"
	"user-svc/internal/domain/errs"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sqlx.DB {
	// Load configuration from config file
	cfg, err := config.LoadConfig("../../../config.yaml")
	require.NoError(t, err, "Failed to load config")
	require.NotNil(t, cfg)

	// Build connection string from config
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err, "Failed to connect to test database")

	// Configure connection pool from config
	if cfg.Database.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	}
	if cfg.Database.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	}

	// Clean up the users table before each test
	_, err = db.Exec("DELETE FROM users")
	require.NoError(t, err)

	return db
}

func TestUserRepository_CreateAndGetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Test data
	email := "test@example.com"
	username := "testuser"
	passwordHash := "hashedpassword123"

	// Create a new user
	user, err := models.NewUser(email, passwordHash, username)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Test Create
	err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// Test GetByEmail
	retrievedUser, err := repo.GetByEmail(ctx, email)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedUser)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Email, retrievedUser.Email)
	assert.Equal(t, user.Username, retrievedUser.Username)
	assert.Equal(t, user.PasswordHash, retrievedUser.PasswordHash)
}

func TestUserRepository_CreateAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Test data
	email := "test2@example.com"
	username := "testuser2"
	passwordHash := "hashedpassword456"

	// Create a new user
	user, err := models.NewUser(email, passwordHash, username)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Test Create
	err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// Test GetByID
	retrievedUser, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedUser)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Email, retrievedUser.Email)
	assert.Equal(t, user.Username, retrievedUser.Username)
	assert.Equal(t, user.PasswordHash, retrievedUser.PasswordHash)
}

func TestUserRepository_DeleteAndVerifyNil(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Test data
	email := "test3@example.com"
	username := "testuser3"
	passwordHash := "hashedpassword789"

	// Create a new user
	user, err := models.NewUser(email, passwordHash, username)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Test Create
	err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// Verify user exists
	retrievedUser, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedUser)

	// Test Delete
	err = repo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// Verify user is deleted - GetByID should return ErrUserNotFound
	deletedUser, err := repo.GetByID(ctx, user.ID)
	assert.Error(t, err)
	assert.Nil(t, deletedUser)
	assert.Equal(t, errs.ErrUserNotFound, err)

	// Verify user is deleted - GetByEmail should return ErrUserNotFound
	deletedUserByEmail, err := repo.GetByEmail(ctx, email)
	assert.Error(t, err)
	assert.Nil(t, deletedUserByEmail)
	assert.Equal(t, errs.ErrUserNotFound, err)
}

func TestUserRepository_CompleteFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Test data
	email := "complete@example.com"
	username := "completeuser"
	passwordHash := "hashedpasswordcomplete"

	// Step 1: Create a user with email
	user, err := models.NewUser(email, passwordHash, username)
	require.NoError(t, err)
	require.NotNil(t, user)

	err = repo.Create(ctx, user)
	assert.NoError(t, err, "Failed to create user")

	// Step 2: Get by email
	retrievedByEmail, err := repo.GetByEmail(ctx, email)
	assert.NoError(t, err, "Failed to get user by email")
	assert.NotNil(t, retrievedByEmail, "User should not be nil when retrieved by email")
	assert.Equal(t, user.ID, retrievedByEmail.ID)
	assert.Equal(t, user.Email, retrievedByEmail.Email)
	assert.Equal(t, user.Username, retrievedByEmail.Username)
	assert.Equal(t, user.PasswordHash, retrievedByEmail.PasswordHash)

	// Step 3: Get by ID
	retrievedByID, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err, "Failed to get user by ID")
	assert.NotNil(t, retrievedByID, "User should not be nil when retrieved by ID")
	assert.Equal(t, user.ID, retrievedByID.ID)
	assert.Equal(t, user.Email, retrievedByID.Email)
	assert.Equal(t, user.Username, retrievedByID.Username)
	assert.Equal(t, user.PasswordHash, retrievedByID.PasswordHash)

	// Step 4: Delete the user
	err = repo.Delete(ctx, user.ID)
	assert.NoError(t, err, "Failed to delete user")

	// Step 5: Verify result is nil (user not found)
	deletedUserByID, err := repo.GetByID(ctx, user.ID)
	assert.Error(t, err, "Should return error when getting deleted user by ID")
	assert.Nil(t, deletedUserByID, "Deleted user should be nil")
	assert.Equal(t, errs.ErrUserNotFound, err)

	deletedUserByEmail, err := repo.GetByEmail(ctx, email)
	assert.Error(t, err, "Should return error when getting deleted user by email")
	assert.Nil(t, deletedUserByEmail, "Deleted user should be nil")
	assert.Equal(t, errs.ErrUserNotFound, err)
}

func TestUserRepository_GetByEmailNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Try to get a non-existent user by email
	user, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, errs.ErrUserNotFound, err)
}

func TestUserRepository_GetByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Try to get a non-existent user by ID
	user, err := repo.GetByID(ctx, uuid.Nil)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, errs.ErrUserNotFound, err)
}

func TestUserRepository_DeleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Try to delete a non-existent user
	err := repo.Delete(ctx, uuid.Nil)
	assert.Error(t, err)
	assert.Equal(t, errs.ErrUserNotFound, err)
}
