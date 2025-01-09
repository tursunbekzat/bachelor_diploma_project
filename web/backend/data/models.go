package data

import (
    "time"
)

// User represents a player in the game
type User struct {
    ID        int       `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`
    CreatedAt time.Time `json:"created_at"`
}

// Game represents a game session
type Game struct {
    ID        int       `json:"id"`
    GameName  string    `json:"game_name"`
    CreatorID int       `json:"creator_id"`
    Status    string    `json:"status"` // waiting, in_progress, finished
    CreatedAt time.Time `json:"created_at"`
}

// Player represents a player in a specific game
type Player struct {
    ID         int        `json:"id"`
    GameID     int        `json:"game_id"`
    UserID     int        `json:"user_id"`
    Role       string        `json:"role_"`
    Health     int        `json:"health"`
    Character  string        `json:"character"`
}

// Role represents a role in the game
type Role struct {
    Name       string `json:"name"`
    Definition string `json:"definition"`
}

// Character represents a character in the game
type Character struct {
    Name   string `json:"name"`
    Definition string `json:"definition"`
    Health int    `json:"health"`
}