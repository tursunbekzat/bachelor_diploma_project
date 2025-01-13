package handlers

import (
	"backend/data"
	"backend/db"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// RegisterUserHandler handles user registration
func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
    var userRequest struct {
        Username string `json:"username"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    err := json.NewDecoder(r.Body).Decode(&userRequest)
    if err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Check if the user already exists
    existingUser, err := db.GetUserByUsername(userRequest.Username)
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    if existingUser != nil {
        http.Error(w, "Username already taken", http.StatusConflict)
        return
    }

    // Create new user
    newUser := &data.User{
        Username:  userRequest.Username,
        Email:     userRequest.Email,
        CreatedAt: time.Now(),
    }

    err = newUser.HashPassword(userRequest.Password)
    if err != nil {
        http.Error(w, "Error hashing password", http.StatusInternalServerError)
        return
    }

    err = db.CreateUser(newUser)
    if err != nil {
        http.Error(w, "Error creating user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(newUser)
}

// RegisterMultipleUsersHandler handles registration of multiple users
func RegisterMultipleUsersHandler(w http.ResponseWriter, r *http.Request) {
    var userRequests []struct {
        Username string `json:"username"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    // Декодируем массив JSON из тела запроса
    err := json.NewDecoder(r.Body).Decode(&userRequests)
    if err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    var createdUsers []data.User
    for _, userRequest := range userRequests {
        // Проверяем, существует ли пользователь
        existingUser, err := db.GetUserByUsername(userRequest.Username)
        if err != nil {
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }
        if existingUser != nil {
            http.Error(w, "Username "+userRequest.Username+" already taken", http.StatusConflict)
            return
        }

        // Создаём нового пользователя
        newUser := &data.User{
            Username:  userRequest.Username,
            Email:     userRequest.Email,
            CreatedAt: time.Now(),
        }

        err = newUser.HashPassword(userRequest.Password)
        if err != nil {
            http.Error(w, "Error hashing password for user "+userRequest.Username, http.StatusInternalServerError)
            return
        }

        err = db.CreateUser(newUser)
        if err != nil {
            http.Error(w, "Error creating user "+userRequest.Username, http.StatusInternalServerError)
            return
        }

        createdUsers = append(createdUsers, *newUser)
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(createdUsers)
}

// LoginHandler handles user login and JWT token generation
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var credentials struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    err := json.NewDecoder(r.Body).Decode(&credentials)
    if err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Fetch user from the database
    user, err := db.GetUserByUsername(credentials.Username)
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    if user == nil {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    // Check password
    err = user.CheckPassword(credentials.Password)
    if err != nil {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    // Generate JWT token
    token, err := data.GenerateJWT(user.ID, user.Username)
    if err != nil {
        http.Error(w, "Could not generate token", http.StatusInternalServerError)
        return
    }

    http.SetCookie(w, &http.Cookie{
        Name:     "token",
        Value:    token,
        Expires:  time.Now().Add(24 * time.Hour),
        Path:     "/",
        Secure:   true,               // Требуется для SameSite=None
        HttpOnly: true,
        SameSite: http.SameSiteNoneMode, // Разрешаем отправку cookie в кросс-доменных запросах
    })

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"token": token})
}

//  GetUsersHandler handles all users who's registered
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
    users, err := db.GetAllUsers()
    if err != nil {
        log.Println("GetAllUsersHandler error!")
        http.Error(w, "Could not retrieve users", http.StatusInternalServerError)
        return
    }
    if users == nil {
        log.Println("Currently No Users!")
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(users)
}

// GetCurrentUser возвращает информацию о текущем пользователе
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Получаем токен из cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
		return
	}

	// Проверяем токен
	claims, err := data.ValidateJWT(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	// Получаем пользователя из базы данных
	user, err := db.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "Server error: Could not retrieve user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Удаляем пароль перед отправкой данных пользователю
	user.Password = ""
    
    // Форматируем дату в ISO 8601
    user.CreatedAt = user.CreatedAt.UTC()


	// Отправляем информацию о пользователе в ответе
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}