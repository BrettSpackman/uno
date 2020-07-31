package main

import (
	"testing"

	"github.com/jak103/uno/db"
	"github.com/jak103/uno/model"
	"github.com/stretchr/testify/assert"
)

// This function is meant to get a game and a player into the data base in a usable state for testing.
func setupGameWithPlayer(database *db.DB) (*model.Game, *model.Player) {

	game, _ := database.CreateGame()

	player, _ := database.CreatePlayer("Player 1")

	game, _ = database.JoinGame(game.ID, player.ID)

	game.DrawPile = generateShuffledDeck()

	database.SaveGame(*game)

	return game, player
}

// Test that you can get the index for the location of the card that the player holds.
func TestCardFromPlayer(t *testing.T) {

	card := model.Card{Color: "Red", Value: "One"}
	player := model.Player{ID: "ID 1", Name: "Player 1", Cards: []model.Card{card}}

	// Test for a card the user is Guarenteed to have.
	index := cardFromPlayer(&player, &card)

	assert.Equal(t, index, 0)

	//Test a card that the user does not have. Guaranteed this by using invalid values for the card
	index = cardFromPlayer(&player, &model.Card{Color: "orange", Value: "whoops"})

	assert.Equal(t, index, -1)
}

func TestDrawCardHelper(t *testing.T) {

	database, _ := db.GetDb()

	game, player := setupGameWithPlayer(database)

	drawCardHelper(game, player)

	assert.Equal(t, len(player.Cards), 1)
}

func TestDrawCard(t *testing.T) {

	// Test passing in a bogus game id, we should get an error
	game, err := drawCard("Bogus game id", "Bogus player id")

	// Assert that we got an actual err
	assert.NotNil(t, err, "We did not error on a bogus game id")

	// Generate real game in database and real player
	database, _ := db.GetDb()
	game, player := setupGameWithPlayer(database)

	// Test Drawing a card with a full deck and real player
	game, err = drawCard(game.ID, player.ID)
	game, _ = database.LookupGameByID(game.ID)
	player = &game.Players[game.CurrentPlayer]

	// Assert that no error occured, the player has a new card and the draw pile
	// has one less card
	assert.Nil(t, err, "Failed to draw card.")
	assert.Equal(t, 1, len(player.Cards))
	assert.Equal(t, 107, len(game.DrawPile))

	// Move all cards into the discard pile, Empty out the draw pile completely,
	// and test drawing a card. It should resuffle leaving one card on the discard pile
	game.DiscardPile = append(game.DiscardPile, game.DrawPile...)
	game.DrawPile = game.DrawPile[:0]
	lastCard := game.DiscardPile[len(game.DiscardPile)-1]

	database.SaveGame(*game)

	game, err = drawCard(game.ID, player.ID)
	player = &game.Players[game.CurrentPlayer]

	//Assert no error, player has 2 cards from both draw tests,
	// draw is missing three from two draws and one in discard
	// discard has one last remaining card.
	// Assert last card in discard is actually to proper last card
	assert.Nil(t, err, "Failed to draw card.")
	assert.Equal(t, 2, len(player.Cards))
	assert.Equal(t, 105, len(game.DrawPile))
	assert.Equal(t, 1, len(game.DiscardPile))
	assert.Equal(t, lastCard.Color, game.DiscardPile[0].Color)
	assert.Equal(t, lastCard.Value, game.DiscardPile[0].Value)

	// Empty out both discard and draw piles, it should for now put a new deck on the draw pile.
	game.DrawPile = game.DrawPile[:0]
	game.DiscardPile = game.DiscardPile[:1]
	lastCard = game.DiscardPile[len(game.DiscardPile)-1]

	database.SaveGame(*game)

	game, err = drawCard(game.ID, player.ID)
	player = &game.Players[game.CurrentPlayer]

	// Assert no errors, assert player now has 3 cards
	// assert new draw pile with one missing
	// assert discard still has one card
	// Assert last card in discard is actually to proper last card
	assert.Nil(t, err, "Failed to draw card.")
	assert.Equal(t, 3, len(player.Cards))
	assert.Equal(t, 107, len(game.DrawPile))
	assert.Equal(t, 1, len(game.DiscardPile))
	assert.Equal(t, lastCard.Color, game.DiscardPile[0].Color)
	assert.Equal(t, lastCard.Value, game.DiscardPile[0].Value)

	// Create a bogus player with a bogus ID
	otherPlayer := model.Player{ID: " id 2 ", Name: "Name 2", Cards: []model.Card{}}

	// Simulate a someone trying to participate in a game they are not a part of.
	_, err = drawCard(game.ID, otherPlayer.ID)

	// Assert that we got an error from the draw card function as we should have.
	// Assert that the player didn't get any cards
	// Assert that the draw pile didn't lose any cards.
	assert.NotNil(t, err, "Player not in the game drew a card. Please make sure only players in the game can draw")
	assert.Equal(t, "You cannot participate in a game you do not belong", err.Error())
	assert.Equal(t, 0, len(otherPlayer.Cards))
	assert.Equal(t, 107, len(game.DrawPile))

	// Create a real player and add them to the game so there is more than one player.
	player2, _ := database.CreatePlayer("Player 2")

	game, _ = database.JoinGame(game.ID, player2.ID)

	database.SaveGame(*game)

	//Simulate drawing out of turn
	_, err = drawCard(game.ID, player2.ID)

	// Assert that we got an error from the draw card function as we should have.
	// Assert that the player didn't get any cards
	// Assert that the draw pile didn't lose any cards.
	assert.NotNil(t, err, "Player drew out of turn. Please make sure only the player whoes turn it is can play.")
	assert.Equal(t, "It is not your turn to play", err.Error())
	assert.Equal(t, 0, len(player2.Cards))
	assert.Equal(t, 107, len(game.DrawPile))

}

/*func TestPlayCard(t *testing.T) {

	database, _ := db.GetDb()

	game, player := setupGameWithPlayer(database)

	drawCard(game.ID, player.ID)

	game, err := playCard(game.ID, player.ID, player.Cards[0])

	if err == nil {
		player, _ = database.LookupPlayer(player.ID)
		assert.Equal(t, len(player.Cards), 0)
		assert.Equal(t, len(game.DrawPile), 108)
	} else {
		assert.Fail(t, "Failed to play card.")
	}
}*/

// func TestCheckForWinner(t *testing.T) {
// 	players = []string{"player1", "player2"}
// 	allCards = make(map[string][]model.Card)
// 	dealCards()
// 	assert.Equal(t, "", checkForWinner())
// 	allCards[players[0]] = make([]model.Card, 0)
// 	assert.Equal(t, "player1", checkForWinner())
// }


func TestDealCards(t *testing.T) {

	// Test passing in a bogus game id, we should get an error
	game, err := dealCards("Bogus game id")

	// Assert that we got an actual error
	assert.NotNil(t, err, "We did not error on a bogus game id")

	// Generate real game in database and real player
	database, _ := db.GetDb()
	game, player := setupGameWithPlayer(database)

	// Test Drawing a card with a full deck and real player
	game, err = dealCards(game.ID)
	player = &game.Players[game.CurrentPlayer] //getting from the game who the current player is

	// Assert that no error occured, the player has a new card and the draw pile
	// has one less card
	assert.Nil(t, err, "Failed to deal cards.")
	assert.Equal(t, 7, len(player.Cards))
	assert.Equal(t, 101, len(game.DrawPile))
	assert.Equal(t, 1, len(game.DiscardPile))
	
	// Create additional players and add them to the game
	player2, _ := database.CreatePlayer("Player 2")
	player3, _ := database.CreatePlayer("Player 3")
	player4, _ := database.CreatePlayer("Player 4")
	player5, _ := database.CreatePlayer("Player 5")

	game, _ = database.JoinGame(game.ID, player2.ID)
	//Have to save in between each player being added or the game state wont recall any but the last
	database.SaveGame(*game)
	game, _ = database.JoinGame(game.ID, player3.ID)
	database.SaveGame(*game)
	game, _ = database.JoinGame(game.ID, player4.ID)
	database.SaveGame(*game)
	game, _ = database.JoinGame(game.ID, player5.ID)
	database.SaveGame(*game)


	//refresh the drawPile and the discardPile
	game.DrawPile = generateShuffledDeck()
	game.DiscardPile = []model.Card{}

	// Test Drawing a card with a full deck and multiple players
	game, err = dealCards(game.ID)
	// Assert that no error occured, the player has a new card and the draw pile
	// has one less card
	assert.Nil(t, err, "Failed to deal multiple players cards.")
	for _, player := range game.Players {
		assert.Equal(t, 7, len(player.Cards))
		}
	assert.Equal(t, 73, len(game.DrawPile))
	assert.Equal(t, 1, len(game.DiscardPile))
	
	//refresh the drawPile and the discardPile
	game.DrawPile = generateShuffledDeck()
	game.DiscardPile = []model.Card{}
	//remove all but the top 5 cards so we will run out for sure part way through dealing
	game.DrawPile = game.DrawPile[:5]
	database.SaveGame(*game)

	//run deal cards
	game, err = dealCards(game.ID)
	//check if the draw pile refreshed
	assert.Equal(t, 78, len(game.DrawPile))


	//database.SaveGame(*game)		//Use this if you make a modification to the game, drawcard will save it for you but need to call whenever you change game state
}







//TODO add asserts that it returns the proper game
//This is something for us to call so that we can check as needed
func checkGameAgainstDatabase(t *testing.T, game *model.Game) {

}