package main

import (
	"log"
	"net/http"

	"backend/middlewares"
	"backend/handlers"
    "backend/db"

	"github.com/gorilla/mux"
)

func main() {

    db.ConnectDB()

    router := mux.NewRouter()

    // Public routes
    router.HandleFunc("/api/register", handlers.RegisterUserHandler).Methods("POST")
    router.HandleFunc("/api/register_users", handlers.RegisterMultipleUsersHandler).Methods("POST")
    router.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")
    router.HandleFunc("/api/user", handlers.GetCurrentUser).Methods("GET")
	router.HandleFunc("/api/users", handlers.GetUsersHandler).Methods("GET")

    // Protected routes
    protected := router.PathPrefix("/api").Subrouter()
    protected.Use(middlewares.JWTAuthMiddleware)
	protected.HandleFunc("/games", handlers.GetAllGamesHandler).Methods("GET")    // Получение списка всех игр
	protected.HandleFunc("/games/new", handlers.CreateGameHandler).Methods("POST")    // Создание игры
	protected.HandleFunc("/games/{id}", handlers.GetGameDetailsHandler).Methods("GET")   // Получение деталей игры
	protected.HandleFunc("/games/join", handlers.JoinGameHandler).Methods("POST") // Присоединение к игре
	protected.HandleFunc("/games/{id}/start", handlers.StartGameHandler).Methods("POST") // Запуск игры
    protected.HandleFunc("/games/{id}/delete", handlers.DeleteGameHandler).Methods("DELETE")

    log.Println("Server started on :8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}

