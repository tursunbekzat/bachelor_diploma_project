package handlers

import (
	"backend/data"
	"backend/db"
	"backend/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// StartTurnHandler starts the player's turn
func StartTurnHandler(w http.ResponseWriter, r *http.Request) {
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

	gameIDStr := mux.Vars(r)["id"]
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	gameState, err := db.GetGameState(gameID)
	if err != nil {
		http.Error(w, "Could not retrieve game state", http.StatusInternalServerError)
		return
	}

	if gameState.CurrentTurn != claims.UserID {
		http.Error(w, "It's not your turn", http.StatusForbidden)
		return
	}

	for i := 0; i < 2; i++ {
		card, err := db.DrawCard(gameID)
		if err != nil {
			http.Error(w, "Could not draw card", http.StatusInternalServerError)
			return
		}
		err = db.AddCardToPlayerHand(claims.UserID, gameID, card.ID)
		if err != nil {
			http.Error(w, "Could not add card to hand", http.StatusInternalServerError)
			return
		}
	}

	err = db.UpdateGameStatePhase(gameID, "play")
	if err != nil {
		http.Error(w, "Could not update game phase", http.StatusInternalServerError)
		return
	}

	NotifyPlayers(gameID, "turn_started", map[string]interface{}{
		"player_id": claims.UserID,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Turn started. Draw phase complete."})
}

func PlayCardHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем токен и проверяем пользователя
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

	// Получаем ID игры и карту из запроса
	vars := mux.Vars(r)
	gameID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	var cardRequest struct {
		CardID   int `json:"card_id"`
		TargetID int `json:"target_id"`
	}
	err = json.NewDecoder(r.Body).Decode(&cardRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверяем текущее состояние игры
	gameState, err := db.GetGameState(gameID)
	if err != nil {
		http.Error(w, "Could not retrieve game state", http.StatusInternalServerError)
		return
	}
	if gameState == nil {
		http.Error(w, "Game state not found", http.StatusNotFound)
		return
	}

	if gameState.CurrentTurn != claims.UserID {
		http.Error(w, "It's not your turn", http.StatusForbidden)
		return
	}

	if gameState.CurrentPhase != "play" {
		http.Error(w, "You cannot play a card outside of the play phase", http.StatusForbidden)
		return
	}

    // Получаем название карты по ID
    card, err := db.GetCardByID(cardRequest.CardID)
    if err != nil {
        http.Error(w, "Could not retrieve card", http.StatusInternalServerError)
        return
    }

    // Проверяем, есть ли у игрока эта карта на руке
    hasCard, err := db.CheckPlayerHasCard(claims.UserID, gameID, card.Name)
    if err != nil {
        http.Error(w, "Could not check player hand", http.StatusInternalServerError)
        return
    }
    
	if !hasCard {
		http.Error(w, "You don't have this card", http.StatusForbidden)
		return
	}

	// Удаляем карту из руки игрока
	err = db.RemoveCardFromPlayerHand(claims.UserID, gameID, cardRequest.CardID)
	if err != nil {
		http.Error(w, "Could not remove card from hand", http.StatusInternalServerError)
		return
	}

	// Добавляем карту в сброс
	err = db.DiscardCard(gameID, cardRequest.CardID)
	if err != nil {
		http.Error(w, "Could not discard card", http.StatusInternalServerError)
		return
	}

	// Применяем эффект карты
	err = ApplyCardEffect(gameID, claims.UserID, cardRequest.CardID, cardRequest.TargetID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not apply card effect: %v", err), http.StatusInternalServerError)
		return
	}

	// Отправляем уведомление игрокам через WebSocket
	NotifyPlayers(gameID, "card_played", map[string]any{
		"player_id": claims.UserID,
		"card_id":   cardRequest.CardID,
		"target_id": cardRequest.TargetID,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Card played successfully"})
}

// ApplyCardEffect applies the effect of a played card
func ApplyCardEffect(gameID int, userID int, cardID int, targetID int) error {
	card, err := db.GetCardByID(cardID)
	if err != nil {
		return fmt.Errorf("could not retrieve card: %v", err)
	}

	switch card.Name {
	case "Bang!":
		return utils.HandleBangEffect(gameID, userID, targetID)
	case "Missed!":
		return utils.HandleMissedEffect(gameID, userID)
	case "Beer":
		return utils.HandleBeerEffect(gameID, userID)
	case "Jail":
		return utils.HandleJailEffect(gameID, targetID)
	case "Dynamite":
		return utils.HandleDynamiteEffect(gameID, userID)
	case "Barrel":
		return utils.HandleBarrelEffect(gameID, userID)
	default:
		return fmt.Errorf("unknown card effect")
	}
}

// EndTurnHandler ends the player's turn
func EndTurnHandler(w http.ResponseWriter, r *http.Request) {
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

	gameIDStr := mux.Vars(r)["id"]
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	gameState, err := db.GetGameState(gameID)
	if err != nil {
		http.Error(w, "Could not retrieve game state", http.StatusInternalServerError)
		return
	}

	if gameState.CurrentTurn != claims.UserID {
		http.Error(w, "It's not your turn", http.StatusForbidden)
		return
	}

	nextPlayerID, err := db.GetNextPlayerID(gameID, gameState.CurrentTurn)
	if err != nil {
		http.Error(w, "Could not get next player", http.StatusInternalServerError)
		return
	}

	err = db.UpdateGameStateTurn(gameID, nextPlayerID)
	if err != nil {
		http.Error(w, "Could not update game state", http.StatusInternalServerError)
		return
	}

	NotifyPlayers(gameID, "turn_ended", map[string]interface{}{
		"previous_player_id": gameState.CurrentTurn,
		"next_player_id":     nextPlayerID,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Turn ended. Next player's turn."})
}



var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}
var clients = make(map[*websocket.Conn]bool) // Connected clients
var broadcast = make(chan Message)           // Channel for sending messages

type Message struct {
	GameID int         `json:"game_id"`
	Event  string      `json:"event"`
	Data   interface{} `json:"data"`
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// NotifyPlayers sends a notification to all connected players
func NotifyPlayers(gameID int, event string, data interface{}) {
	msg := Message{
		GameID: gameID,
		Event:  event,
		Data:   data,
	}
	broadcast <- msg
}
