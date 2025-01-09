package db

import (
	"backend/data"
	"database/sql"
	"fmt"
)

// CreateUser adds a new user to the database
func CreateUser(user *data.User) error {
    query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id`
    err := DB.QueryRow(query, user.Username, user.Email, user.Password).Scan(&user.ID)
    if err != nil {
        return fmt.Errorf("could not insert user: %v", err)
    }
    return nil
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(username string) (*data.User, error) {
    query := `SELECT id, username, email, password, created_at FROM users WHERE username = $1`
    user := &data.User{}
    err := DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("could not query user: %v", err)
    }
    return user, nil
}

// GetAllUsers retrieves users list
func GetAllUsers() ([]*data.User, error) {
    rows, err := DB.Query("SELECT id, username, email FROM users")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*data.User
    for rows.Next() {
        var user data.User
        if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
            return nil, err
        }
        users = append(users, &user)
    }
    return users, nil
}

// GetCurrentUser retrieves the current user based on the provided user ID
func GetCurrentUser(userID int) (*data.User, error) {
    query := `SELECT id, username, email, created_at FROM users WHERE id = $1`
    user := &data.User{}
    err := DB.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("could not retrieve current user: %v", err)
    }
    return user, nil
}