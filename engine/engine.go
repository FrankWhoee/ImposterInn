package engine

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strconv"
)

type Engine struct {
	GameState *GameState
}

func CreateOrderedDeck() []Card {
	deck := []Card{}

	for i := 0; i < 6; i++ {
		deck = append(deck, King)
	}

	for i := 0; i < 6; i++ {
		deck = append(deck, Queen)
	}

	for i := 0; i < 6; i++ {
		deck = append(deck, Ace)
	}

	for i := 0; i < 2; i++ {
		deck = append(deck, Joker)
	}

	return deck
}

func ShuffleDeck(deck []Card) {
	for i := range deck {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
}

func NewEngine() *Engine {
	e := new(Engine)
	e.GameState = new(GameState)
	e.ResetRound()
	return e
}

func (e *Engine) ResetRound() {
	g := e.GameState
	g.TableCard = Card(rand.Intn(3))
	g.CardsLastPlayed = []Card{}
	g.TurnHistory = []Turn{}
	deck := CreateOrderedDeck()
	ShuffleDeck(deck)

	for i := 0; i < 4; i++ {
		if i >= len(g.Players) {
			g.Players = append(g.Players, new(Player))
			g.Players[i].Id = i
			g.Players[i].CurrentCartridge = 0
			g.Players[i].LiveCartridge = 1 + rand.Intn(6)
		}
		g.Players[i].Cards = []Card{}
		for j := i * 5; j < (i+1)*5; j++ {
			g.Players[i].Cards = append(g.Players[i].Cards, deck[j])
		}
	}

	g.CurrentPlayerId = rand.Intn(int(len(g.Players)))
	e.nextPlayer()
}

func remove(s []Card, c Card) []Card {
	i := slices.Index(s, c)
	return slices.Delete(s, i, i+1)
}

func (e *Engine) nextPlayer() {
	if e.numPlayersParticipating() <= 0 {
		panic("cannot find next player: there are no more players remaining.")
	}
	gameState := e.GameState
	gameState.CurrentPlayerId++
	gameState.CurrentPlayerId = gameState.CurrentPlayerId % 4
	for !gameState.CurrentPlayer().IsAlive() || len(gameState.CurrentPlayer().Cards) <= 0 {
		gameState.CurrentPlayerId++
		gameState.CurrentPlayerId = gameState.CurrentPlayerId % 4
	}
}

// interp. Representation of the results of a turn.
// ChallengeResult: The result of a challenge if a challenge was issued
// E: Any errors that occured in the processing of the turn
type PlayResult struct {
	ChallengeResult *ChallengeResult
	E               error
}

// interp. Representation of the results of a challenge.
// ChallengePassed: (A challenge was issued this turn) && (The challenge was passed, i.e. the challenged player did not lie)
// CardReveal: The cards that the challenged player played last round
// TriggerPlayer: The player that has to pull the trigger
type ChallengeResult struct {
	ChallengePassed bool
	CardReveal      []Card
	TriggerPlayer   *Player
}

func (e *Engine) Winner() *Player {
	var winner *Player
	for _, p := range e.GameState.Players {
		if p.IsAlive() {
			if winner == nil {
				winner = p
			} else {
				return nil
			}

		}
	}
	return winner
}

func (e *Engine) numPlayersParticipating() int {
	count := 0
	for _, p := range e.GameState.Players {
		if p.IsAlive() && len(p.Cards) > 0 {
			count++
		}
	}
	return count
}

// interp. Processes a turn and updates the gamestate and returns the results of that turn
func (e *Engine) Play(t Turn) PlayResult {
	gameState := e.GameState
	CurrentPlayer := gameState.CurrentPlayer()
	PreviousPlayer := gameState.PreviousPlayer()
	if CurrentPlayer.CurrentCartridge >= CurrentPlayer.LiveCartridge {
		return PlayResult{nil, fmt.Errorf("invalid turn: Player %d is dead", CurrentPlayer.Id)}
	}
	if t.Action == Play {
		// interp. Cards played that are invalid (i.e. cards that the player wants to play but doesn't have)
		badCards := []Card{}
		// interp. what the player's hand will look like after this turn
		// This variable will eventually be that
		newPlayerCards := make([]Card, len(CurrentPlayer.Cards))
		copy(newPlayerCards, CurrentPlayer.Cards)

		// Remove played cards
		for _, c := range t.Cards {
			if slices.Contains(CurrentPlayer.Cards, c) {
				newPlayerCards = remove(newPlayerCards, c)	
			} else {
				badCards = append(badCards, c)
			}
		}
		if len(badCards) > 0 {
			return PlayResult{nil, errors.New("Played cards that didn't exist: " + CardListToString(badCards))}
		}
		CurrentPlayer.Cards = newPlayerCards

		if e.numPlayersParticipating() == 1 {
			e.nextPlayer()
			return e.Play(Turn{Action: Challenge})
		} else {
			e.GameState.TurnHistory = append(e.GameState.TurnHistory, t)
			gameState.CardsLastPlayed = t.Cards
			gameState.PreviousPlayerId = gameState.CurrentPlayerId
			e.nextPlayer()
			return PlayResult{nil, nil}
		}
	} else if t.Action == Challenge {
		cardsLastPlayed := gameState.CardsLastPlayed
		if len(gameState.CardsLastPlayed) <= 0 {
			return PlayResult{nil, errors.New("invalid challenge: No cards have been played yet")}
		}
		for _, c := range gameState.CardsLastPlayed {
			if c != Joker && c != gameState.TableCard {
				PreviousPlayer.CurrentCartridge++
				e.ResetRound()
				cr := new(ChallengeResult)
				cr.ChallengePassed = false
				cr.CardReveal = cardsLastPlayed
				cr.TriggerPlayer = PreviousPlayer
				return PlayResult{cr, nil}
			}
		}

		CurrentPlayer.CurrentCartridge++
		e.ResetRound()
		cr := new(ChallengeResult)
		cr.ChallengePassed = true
		cr.CardReveal = cardsLastPlayed
		cr.TriggerPlayer = CurrentPlayer
		return PlayResult{cr, nil}
	} else {
		return PlayResult{nil, errors.New("Unknown move: " + strconv.Itoa(int(t.Action)))}
	}
}
