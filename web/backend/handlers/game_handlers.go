package handlers

import (
	"backend/data"
	"backend/db"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)



// CreateGameHandler handles the creation of a new game
func CreateGameHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("token")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    claims, err := data.ValidateJWT(cookie.Value)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    var gameRequest struct {
        GameName string `json:"game_name"`
    }

    err = json.NewDecoder(r.Body).Decode(&gameRequest)
    if err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    newGame := &data.Game{
        GameName:  gameRequest.GameName,
        CreatorID: claims.UserID,
        Status:   "waiting",
        CreatedAt: time.Now(),
    }

    err = db.CreateGame(newGame)
    if err != nil {
        http.Error(w, "Could not create game", http.StatusInternalServerError)
        return
    }
    log.Println(newGame.GameName, "- Game Created!")

    // Add the creator to the game automatically
    err = db.AddPlayerToGame(newGame.ID, claims.UserID)
    if err != nil {
        http.Error(w, "Could not add creator to game", http.StatusInternalServerError)
        return
    }
    log.Println(claims.Username, "automatically added to the Game")

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(newGame)
}

// GetAllGamesHandler возвращает список всех игр
func GetAllGamesHandler(w http.ResponseWriter, r *http.Request) {
    games, err := db.GetAllGames()
    if err != nil {
        log.Println("GetAllGamesHandler error!")
        http.Error(w, "Could not retrieve games", http.StatusInternalServerError)
        return
    }
    if games == nil {
        log.Println("Currently No Games!")
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(games)
}

// GetGameDetailsHandler returns detailed information about a specific game
func GetGameDetailsHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("token")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    claims, err := data.ValidateJWT(cookie.Value)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }
    log.Println("GetGameDetailsHandler: ", claims)

    // Получаем параметр {id} из пути
    vars := mux.Vars(r)
    gameIDStr := vars["id"]

    gameID, err := strconv.Atoi(gameIDStr)
    if err != nil {
        http.Error(w, "Invalid game ID", http.StatusBadRequest)
        return
    }

    game, err := db.GetGameByID(gameID)
    if err != nil {
        http.Error(w, "Could not retrieve game", http.StatusInternalServerError)
        return
    }
    log.Println("Current game -", game.GameName)

    if game == nil {
        http.Error(w, "Game not found", http.StatusNotFound)
        return
    }

    players, err := db.GetPlayersInGame(gameID)
    if err != nil {
        http.Error(w, "Could not retrieve players", http.StatusInternalServerError)
        return
    }
    log.Println("Players in game -", players)

    gameDetails := struct {
        Game     *data.Game     `json:"game"`
        Players  []*data.Player `json:"players"`
    }{
        Game:    game,
        Players: players,
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(gameDetails)
}

// JoinGameHandler handles players joining an existing game
func JoinGameHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("token")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    claims, err := data.ValidateJWT(cookie.Value)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    var joinRequest struct {
        GameID int `json:"game_id"`
    }

    err = json.NewDecoder(r.Body).Decode(&joinRequest)
    if err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    game, err := db.GetGameByID(joinRequest.GameID)
    if err != nil {
        http.Error(w, "Could not retrieve game", http.StatusInternalServerError)
        return
    }

    if game == nil {
        http.Error(w, "Game not found", http.StatusNotFound)
        return
    }

    err = db.AddPlayerToGame(joinRequest.GameID, claims.UserID)
    if err != nil {
        http.Error(w, "Could not join game", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Joined game successfully"})
}

// StartGameHandler handles the logic to start a new game
func StartGameHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("token")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    claims, err := data.ValidateJWT(cookie.Value)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }
    log.Println("Start game:", claims)

    vars := mux.Vars(r)
    gameIDStr := vars["id"]

    gameID, err := strconv.Atoi(gameIDStr)
    if err != nil {
        http.Error(w, "Invalid game ID", http.StatusBadRequest)
        return
    }

    game, err := db.GetGameByID(gameID)
    if err != nil {
        http.Error(w, "Could not retrieve game", http.StatusInternalServerError)
        return
    }

    if game == nil {
        http.Error(w, "Game not found", http.StatusNotFound)
        return
    }

    players, err := db.GetPlayersInGame(gameID)
    if err != nil {
        http.Error(w, "Could not retrieve players", http.StatusInternalServerError)
        return
    }

    if len(players) < 4 {
        http.Error(w, "Not enough players to start the game", http.StatusBadRequest)
        return
    }

    err = AssignRolesAndCharacters(gameID)
    if err != nil {
        http.Error(w, "Could not assign roles and characters", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Game started successfully"})
}

// AssignRolesAndCharacters assigns roles and characters to players based on the number of players
func AssignRolesAndCharacters(gameID int) error {
    players, err := db.GetPlayersInGame(gameID)
    if err != nil {
        log.Println("Error in Getting Players List")
        return err
    }

    if len(players) == 0 {
        log.Println("no players found in game")
        return fmt.Errorf("no players found in game")
    }

    numPlayers := len(players)
    if numPlayers < 4 || numPlayers > 7 {
        log.Println("Error: invalid number of players")
        return fmt.Errorf("invalid number of players: %d", numPlayers)
    }

    characters, err := db.GetAvailableCharacters(gameID)
    if err != nil {
        log.Println("Error getting characters")
        return err
    }
    log.Println("Characters:", len(characters))

    roles, err := db.GetRolesByPlayerCount(numPlayers)
    if err != nil {
        log.Println("Error getting roles")
        return err
    }
    log.Println("Roles:", len(roles))

    if len(roles) < len(players) || len(characters) < len(players) {
        log.Println("not enough roles or characters for all players")
        return fmt.Errorf("not enough roles or characters for all players")
    }

    rand.Shuffle(len(roles), func(i, j int) { roles[i], roles[j] = roles[j], roles[i] })
    rand.Shuffle(len(characters), func(i, j int) { characters[i], characters[j] = characters[j], characters[i] })

    for i, player := range players {
        role := roles[i]
        character := characters[i]

        // Шериф получает +1 к здоровью
        health := character.Health
        if role.Name == "Sheriff" {
            health++
        }

        err := db.UpdatePlayerRoleAndCharacter(player.ID, role.Name, character.Name, health)
        if err != nil {
            log.Println("Error UpdatePlayerRoleAndCharacter")
            return err
        }
    }

    return nil
}

// DeleteGameHandler handles the deletion of a game
func DeleteGameHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("token")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    claims, err := data.ValidateJWT(cookie.Value)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    vars := mux.Vars(r)
    gameIDStr := vars["id"]

    gameID, err := strconv.Atoi(gameIDStr)
    if err != nil {
        http.Error(w, "Invalid game ID", http.StatusBadRequest)
        return
    }

    game, err := db.GetGameByID(gameID)
    if err != nil {
        http.Error(w, "Could not retrieve game", http.StatusInternalServerError)
        return
    }

    if game == nil {
        http.Error(w, "Game not found", http.StatusNotFound)
        return
    }

    if game.CreatorID != claims.UserID {
        http.Error(w, "Only the creator can delete the game", http.StatusForbidden)
        return
    }

    err = db.DeleteGame(gameID)
    if err != nil {
        http.Error(w, "Could not delete game", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Game deleted successfully"})
}