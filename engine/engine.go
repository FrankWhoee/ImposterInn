package engine

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"strings"
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

// interp. Representation of the results of a challenge.
// ChallengePassed: (A challenge was issued this turn) && (The challenge was passed, i.e. the challenged player did not lie)
// CardReveal: The cards that the challenged player played last round
// TriggerPlayer: The player that has to pull the trigger
type ChallengeResult struct {
	ChallengePassed  bool
	CardReveal       []Card
	ChallengerPlayer *Player
	ChallengedPlayer *Player
}

func (cr *ChallengeResult) ToIIP() string {
	var sb strings.Builder

	passBit := 0
	if cr.ChallengePassed {
		passBit = 1
	}
	// interp. {cr.ChallengerPlayer.Id} challenged {cr.ChallengedPlayer}. Did this challenge pass: {passBit}. Number of cards played last turn: {len(cr.CardReveal)}
	sb.WriteString(fmt.Sprintf("chre %d %d %d %d", cr.ChallengerPlayer.Id, cr.ChallengedPlayer, passBit, len(cr.CardReveal)))
	for i,c := range cr.CardReveal {
		sb.WriteString(fmt.Sprintf("care%d %d", i, int(c)))
	}
	return sb.String()
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
func (e *Engine) Play(t Turn) (*ChallengeResult, error) {
	gameState := e.GameState
	CurrentPlayer := gameState.CurrentPlayer()
	PreviousPlayer := gameState.PreviousPlayer()
	if CurrentPlayer.CurrentCartridge >= CurrentPlayer.LiveCartridge {
		return nil, fmt.Errorf("invalid turn: Player %d is dead", CurrentPlayer.Id)
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
			return nil, errors.New("Played cards that didn't exist: " + CardListToString(badCards))
		}
		CurrentPlayer.Cards = newPlayerCards

		if e.numPlayersParticipating() == 1 {
			// e.nextPlayer()
			return e.Play(Turn{Action: Challenge, PlayerId: CurrentPlayer.Id})
		} else {
			e.GameState.TurnHistory = append(e.GameState.TurnHistory, t)
			gameState.CardsLastPlayed = t.Cards
			gameState.PreviousPlayerId = gameState.CurrentPlayerId
			e.nextPlayer()
			return nil, nil
		}
	} else if t.Action == Challenge {
		cardsLastPlayed := gameState.CardsLastPlayed
		if len(gameState.CardsLastPlayed) <= 0 {
			return nil, errors.New("invalid challenge: No cards have been played yet")
		}
		for _, c := range gameState.CardsLastPlayed {
			if c != Joker && c != gameState.TableCard {
				// Card that does not match TableCard found, so the challenge succeeded (i.e challenged player must pull the trigger)
				PreviousPlayer.CurrentCartridge++
				e.ResetRound()
				cr := new(ChallengeResult)
				cr.ChallengePassed = false
				cr.CardReveal = cardsLastPlayed
				cr.ChallengerPlayer = CurrentPlayer
				cr.ChallengedPlayer = PreviousPlayer
				return cr, nil
			}
		}

		// All cards match table card, so the challenge failed (i.e the challenger must pull the trigger)
		CurrentPlayer.CurrentCartridge++
		e.ResetRound()
		cr := new(ChallengeResult)
		cr.ChallengePassed = true
		cr.CardReveal = cardsLastPlayed
		cr.ChallengerPlayer = CurrentPlayer
		cr.ChallengedPlayer = PreviousPlayer
		return cr, nil
	} else {
		return nil, errors.New("Unknown move: " + strconv.Itoa(int(t.Action)))
	}
}
