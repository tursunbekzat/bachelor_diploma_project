package main

import (
	"log"
	"net/http"

	"backend/db"
	"backend/handlers"
	"backend/middlewares"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	hand "github.com/gorilla/handlers"
)

// Инициализация WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// Подключение к базе данных
	db.ConnectDB()

	// Создание маршрутизатора
	router := mux.NewRouter()

	// Public routes (доступны без аутентификации)
	router.HandleFunc("/api/register", handlers.RegisterUserHandler).Methods("POST")
	router.HandleFunc("/api/register_users", handlers.RegisterMultipleUsersHandler).Methods("POST")
	router.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/api/users", handlers.GetUsersHandler).Methods("GET")

	// Protected routes (требуют аутентификации)
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middlewares.JWTAuthMiddleware)


	protected.HandleFunc("/user", handlers.GetCurrentUser).Methods("GET") 

	// Обработчики для игр
	protected.HandleFunc("/games", handlers.GetAllGamesHandler).Methods("GET")             // Получение списка всех игр
	protected.HandleFunc("/games/new", handlers.CreateGameHandler).Methods("POST")         // Создание новой игры
	protected.HandleFunc("/games/{id}", handlers.GetGameDetailsHandler).Methods("GET")     // Получение деталей игры
	protected.HandleFunc("/games/join", handlers.JoinGameHandler).Methods("POST")          // Присоединение к игре
	protected.HandleFunc("/games/{id}/start", handlers.StartGameHandler).Methods("POST")   // Запуск игры
	protected.HandleFunc("/games/{id}/delete", handlers.DeleteGameHandler).Methods("DELETE") // Удаление игры
	// Обработчики игрового процесса
	protected.HandleFunc("/games/{id}/play", handlers.PlayCardHandler).Methods("POST")     // Разыгрывание карты
	protected.HandleFunc("/games/{id}/end", handlers.EndTurnHandler).Methods("POST")       // Завершение хода

	// WebSocket route
	protected.HandleFunc("/ws", handlers.WebSocketHandler).Methods("GET")

	// Enable CORS
	corsOptions := hand.CORS(
		hand.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		hand.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		hand.AllowedOrigins([]string{"http://localhost:5173"}),
		hand.AllowCredentials(), // Разрешаем учетные данные (cookie)
	)

	// Запуск сервера
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsOptions(router)))

}
