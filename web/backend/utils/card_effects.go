package utils

import (
	"backend/db"
	"fmt"
	"math/rand"
	"time"
)

// Объявляем канал для отправки сообщений через WebSocket
var broadcast = make(chan map[string]interface{})

// HandleBangEffect handles the effect of the Bang! card
func HandleBangEffect(gameID int, userID int, targetID int) error {
	// Проверяем, есть ли у цели карта Missed!
	hasMissed, missedCardID, err := db.CheckPlayerHasCardByName(targetID, gameID, "Missed!")
	if err != nil {
		return fmt.Errorf("could not check target's hand: %v", err)
	}

	if hasMissed {
		// Цель использует карту Missed!
		err = db.RemoveCardFromPlayerHand(targetID, gameID, missedCardID)
		if err != nil {
			return fmt.Errorf("could not remove Missed! card: %v", err)
		}
		NotifyPlayers(gameID, "card_effect", map[string]interface{}{
			"player_id":  userID,
			"target_id":  targetID,
			"effect":     "Missed!",
			"successful": true,
		})
		return nil // Атака нейтрализована
	}

	// Если у цели нет карты Missed!, она теряет 1 здоровье
	err = db.DecreasePlayerHealth(targetID)
	if err != nil {
		return fmt.Errorf("could not decrease target's health: %v", err)
	}

	NotifyPlayers(gameID, "card_effect", map[string]interface{}{
		"player_id": userID,
		"target_id": targetID,
		"effect":    "Bang!",
		"damage":    1,
	})

	return nil
}

// HandleMissedEffect handles the effect of the Missed! card
func HandleMissedEffect(gameID int, userID int) error {
	// В текущей реализации эффект Missed! уже обрабатывается в HandleBangEffect
	return nil
}

// HandleBeerEffect handles the effect of the Beer card
func HandleBeerEffect(gameID int, userID int) error {
	err := db.IncreasePlayerHealth(userID)
	if err != nil {
		return fmt.Errorf("could not increase player's health: %v", err)
	}

	NotifyPlayers(gameID, "card_effect", map[string]interface{}{
		"player_id": userID,
		"effect":    "Beer",
		"heal":      1,
	})

	return nil
}

// HandleJailEffect handles the effect of the Jail card
func HandleJailEffect(gameID int, targetID int) error {
	jailCardID, err := db.GetCardIDByName("Jail")
	if err != nil {
		return fmt.Errorf("could not get Jail card ID: %v", err)
	}

	err = db.AddCardToPlayerBoard(targetID, gameID, jailCardID)
	if err != nil {
		return fmt.Errorf("could not place Jail on player's board: %v", err)
	}

	NotifyPlayers(gameID, "card_effect", map[string]interface{}{
		"target_id": targetID,
		"effect":    "Jail",
	})

	return nil
}

// HandleDynamiteEffect handles the effect of the Dynamite card
func HandleDynamiteEffect(gameID int, userID int) error {
	rand.Seed(time.Now().UnixNano())
	chance := rand.Intn(100)

	if chance < 16 { // 1 из 6 шансов, что динамит взорвется
		err := db.DecreasePlayerHealth(userID)
		if err != nil {
			return fmt.Errorf("could not decrease player's health: %v", err)
		}

		NotifyPlayers(gameID, "card_effect", map[string]interface{}{
			"player_id": userID,
			"effect":    "Dynamite",
			"damage":    3,
		})
	} else {
		// Если динамит не взорвался, передаем его следующему игроку
		nextPlayerID, err := db.GetNextPlayerID(gameID, userID)
		if err != nil {
			return fmt.Errorf("could not get next player: %v", err)
		}

		dynamiteCardID, err := db.GetCardIDByName("Dynamite")
		if err != nil {
			return fmt.Errorf("could not get Dynamite card ID: %v", err)
		}

		err = db.AddCardToPlayerBoard(nextPlayerID, gameID, dynamiteCardID)
		if err != nil {
			return fmt.Errorf("could not pass Dynamite to next player: %v", err)
		}

		NotifyPlayers(gameID, "card_effect", map[string]interface{}{
			"player_id":    userID,
			"next_player":  nextPlayerID,
			"effect":       "Dynamite",
			"passed_along": true,
		})
	}

	return nil
}

// HandleBarrelEffect handles the effect of the Barrel card
func HandleBarrelEffect(gameID int, userID int) error {
	rand.Seed(time.Now().UnixNano())
	chance := rand.Intn(100)

	if chance < 50 { // 50% шанс, что игрок избежит выстрела
		NotifyPlayers(gameID, "card_effect", map[string]interface{}{
			"player_id":  userID,
			"effect":     "Barrel",
			"successful": true,
		})
		return nil
	}

	NotifyPlayers(gameID, "card_effect", map[string]interface{}{
		"player_id":  userID,
		"effect":     "Barrel",
		"successful": false,
	})

	return nil
}

// NotifyPlayers отправляет уведомление всем подключенным игрокам через WebSocket
func NotifyPlayers(gameID int, event string, data map[string]interface{}) {
	message := map[string]interface{}{
		"game_id": gameID,
		"event":   event,
		"data":    data,
	}

	// Проверяем, есть ли канал для трансляции
	if broadcast != nil {
		broadcast <- message
	}
}
