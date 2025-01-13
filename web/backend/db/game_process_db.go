package db

import (
	"backend/data"
	"database/sql"
	"fmt"
)


// GenerateDeck fills the deck for a new game based on card copies
func GenerateDeck(gameID int) error {
	query := `SELECT id, copies FROM cards`
	rows, err := DB.Query(query)
	if err != nil {
		return fmt.Errorf("could not query cards: %v", err)
	}
	defer rows.Close()

	var position int
	for rows.Next() {
		var cardID, copies int
		if err := rows.Scan(&cardID, &copies); err != nil {
			return fmt.Errorf("could not scan card: %v", err)
		}

		for i := 0; i < copies; i++ {
			position++
			_, err := DB.Exec(`INSERT INTO deck (game_id, card_id, position) VALUES ($1, $2, $3)`, gameID, cardID, position)
			if err != nil {
				return fmt.Errorf("could not insert card into deck: %v", err)
			}
		}
	}

	return nil
}

// DrawCard draws a card from the deck for a specific game
func DrawCard(gameID int) (*data.Card, error) {
	query := `
		SELECT d.card_id, c.name, c.type, c.description 
		FROM deck d
		JOIN cards c ON d.card_id = c.id
		WHERE d.game_id = $1
		ORDER BY d.position ASC
		LIMIT 1`

	var card data.Card
	err := DB.QueryRow(query, gameID).Scan(&card.ID, &card.Name, &card.Type, &card.Description)
	if err != nil {
		return nil, fmt.Errorf("could not draw card: %v", err)
	}

	// Remove the drawn card from the deck
	_, err = DB.Exec(`DELETE FROM deck WHERE game_id = $1 AND card_id = $2`, gameID, card.ID)
	if err != nil {
		return nil, fmt.Errorf("could not remove card from deck: %v", err)
	}

	return &card, nil
}

// DiscardCard adds a card to the discard pile
func DiscardCard(gameID int, cardID int) error {
	query := `INSERT INTO discard_pile (game_id, card_id) VALUES ($1, $2)`
	_, err := DB.Exec(query, gameID, cardID)
	if err != nil {
		return fmt.Errorf("could not discard card: %v", err)
	}
	return nil
}

// ShuffleDiscardIntoDeck moves the discard pile back into the deck and shuffles it
func ShuffleDiscardIntoDeck(gameID int) error {
	query := `SELECT card_id FROM discard_pile WHERE game_id = $1`
	rows, err := DB.Query(query, gameID)
	if err != nil {
		return fmt.Errorf("could not query discard pile: %v", err)
	}
	defer rows.Close()

	var position int
	for rows.Next() {
		var cardID int
		if err := rows.Scan(&cardID); err != nil {
			return fmt.Errorf("could not scan discard pile: %v", err)
		}

		position++
		_, err := DB.Exec(`INSERT INTO deck (game_id, card_id, position) VALUES ($1, $2, $3)`, gameID, cardID, position)
		if err != nil {
			return fmt.Errorf("could not insert card into deck: %v", err)
		}
	}

	// Clear the discard pile
	_, err = DB.Exec(`DELETE FROM discard_pile WHERE game_id = $1`, gameID)
	if err != nil {
		return fmt.Errorf("could not clear discard pile: %v", err)
	}

	return nil
}

// GetGameState retrieves the current game state for a specific game
func GetGameState(gameID int) (*data.GameState, error) {
	query := `SELECT current_turn, current_phase FROM game_state WHERE game_id = $1`
	gameState := &data.GameState{}
	err := DB.QueryRow(query, gameID).Scan(&gameState.CurrentTurn, &gameState.CurrentPhase)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("could not retrieve game state: %v", err)
	}
	return gameState, nil
}

// AddCardToPlayerHand adds a card to the player's hand
func AddCardToPlayerHand(userID int, gameID int, cardID int) error {
	query := `INSERT INTO player_hand (user_id, game_id, card_id) VALUES ($1, $2, $3)`
	_, err := DB.Exec(query, userID, gameID, cardID)
	if err != nil {
		return fmt.Errorf("could not add card to player hand: %v", err)
	}
	return nil
}

// UpdateGameStatePhase updates the current phase of the game
func UpdateGameStatePhase(gameID int, phase string) error {
	query := `UPDATE game_state SET current_phase = $1 WHERE game_id = $2`
	_, err := DB.Exec(query, phase, gameID)
	if err != nil {
		return fmt.Errorf("could not update game phase: %v", err)
	}
	return nil
}

// RemoveCardFromPlayerHand removes a card from a player's hand
func RemoveCardFromPlayerHand(userID int, gameID int, cardID int) error {
	query := `DELETE FROM player_hand WHERE user_id = $1 AND game_id = $2 AND card_id = $3`
	_, err := DB.Exec(query, userID, gameID, cardID)
	if err != nil {
		return fmt.Errorf("could not remove card from player hand: %v", err)
	}
	return nil
}

// GetCardByID retrieves a card by its ID
func GetCardByID(cardID int) (*data.Card, error) {
	query := `SELECT id, name, type, description FROM cards WHERE id = $1`
	card := &data.Card{}
	err := DB.QueryRow(query, cardID).Scan(&card.ID, &card.Name, &card.Type, &card.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("could not retrieve card: %v", err)
	}
	return card, nil
}

// GetNextPlayerID retrieves the next player's ID in the turn order
func GetNextPlayerID(gameID int, currentPlayerID int) (int, error) {
	query := `SELECT user_id FROM players WHERE game_id = $1 AND user_id > $2 ORDER BY user_id ASC LIMIT 1`
	var nextPlayerID int
	err := DB.QueryRow(query, gameID, currentPlayerID).Scan(&nextPlayerID)
	if err == sql.ErrNoRows {
		// If we've reached the end, return the first player
		query = `SELECT user_id FROM players WHERE game_id = $1 ORDER BY user_id ASC LIMIT 1`
		err = DB.QueryRow(query, gameID).Scan(&nextPlayerID)
	}
	if err != nil {
		return 0, fmt.Errorf("could not get next player: %v", err)
	}
	return nextPlayerID, nil
}

// UpdateGameStateTurn updates the current turn in the game state
func UpdateGameStateTurn(gameID int, nextPlayerID int) error {
	query := `UPDATE game_state SET current_turn = $1 WHERE game_id = $2`
	_, err := DB.Exec(query, nextPlayerID, gameID)
	if err != nil {
		return fmt.Errorf("could not update game turn: %v", err)
	}
	return nil
}

// DecreasePlayerHealth decreases the player's health by 1
func DecreasePlayerHealth(userID int) error {
	query := `UPDATE players SET health = health - 1 WHERE user_id = $1 AND health > 0`
	_, err := DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("could not decrease player's health: %v", err)
	}
	return nil
}

// IncreasePlayerHealth increases the player's health by 1
func IncreasePlayerHealth(userID int) error {
	query := `UPDATE players SET health = health + 1 WHERE user_id = $1`
	_, err := DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("could not increase player's health: %v", err)
	}
	return nil
}

// CheckPlayerHasCard checks if a player has a specific card
func CheckPlayerHasCard(userID int, gameID int, cardName string) (bool, error) {
	query := `
		SELECT 1
		FROM player_hand ph
		JOIN cards c ON ph.card_id = c.id
		WHERE ph.user_id = $1 AND ph.game_id = $2 AND c.name = $3`

	var exists bool
	err := DB.QueryRow(query, userID, gameID, cardName).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("could not check player's hand: %v", err)
	}

	return true, nil
}

// GetCardIDByName retrieves the card ID by its name
func GetCardIDByName(cardName string) (int, error) {
	query := `SELECT id FROM cards WHERE name = $1`
	var cardID int
	err := DB.QueryRow(query, cardName).Scan(&cardID)
	if err != nil {
		return 0, fmt.Errorf("could not retrieve card ID: %v", err)
	}
	return cardID, nil
}

// AddCardToPlayerBoard adds a card to the player's board
func AddCardToPlayerBoard(userID int, gameID int, cardID int) error {
	query := `INSERT INTO player_board (user_id, game_id, card_id) VALUES ($1, $2, $3)`
	_, err := DB.Exec(query, userID, gameID, cardID)
	if err != nil {
		return fmt.Errorf("could not add card to player board: %v", err)
	}
	return nil
}

// CheckPlayerHasCardByName checks if a player has a specific card by its name
func CheckPlayerHasCardByName(userID int, gameID int, cardName string) (bool, int, error) {
	query := `
		SELECT ph.card_id
		FROM player_hand ph
		JOIN cards c ON ph.card_id = c.id
		WHERE ph.user_id = $1 AND ph.game_id = $2 AND c.name = $3`

	var cardID int
	err := DB.QueryRow(query, userID, gameID, cardName).Scan(&cardID)
	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("could not check player's hand: %v", err)
	}

	return true, cardID, nil
}
