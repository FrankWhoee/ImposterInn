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

	g.CurrentPlayer = rand.Intn(int(len(g.Players)))
	e.nextPlayer()
}

func remove(s []Card, c Card) []Card {
	i := slices.Index(s, c)
	return slices.Delete(s, i, i+1)
}

func (e *Engine) nextPlayer() {
	gameState := e.GameState
	gameState.CurrentPlayer++
	gameState.CurrentPlayer = gameState.CurrentPlayer % 4
	for !gameState.Players[gameState.CurrentPlayer].IsAlive() && len(gameState.Players[gameState.CurrentPlayer].Cards) > 0 {
		gameState.CurrentPlayer++
		gameState.CurrentPlayer = gameState.CurrentPlayer % 4
	}
}

type PlayResult struct {
	ChallengePassed bool
	CardReveal      []Card
	TriggerPlayer   *Player
	E               error
}

func (e *Engine) Winner() *Player {
	var winner *Player
	for _,p := range e.GameState.Players {
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

func (e *Engine) AllHandsEmpty() bool {
	for _,p := range e.GameState.Players {
		if p.IsAlive() && len(p.Cards) > 0 {
			return false
		} 
	}
	return true
}

func (e *Engine) Play(t Turn) PlayResult {
	gameState := e.GameState
	CurrentPlayer := gameState.Players[gameState.CurrentPlayer]
	PreviousPlayer := gameState.Players[gameState.PreviousPlayer]
	if CurrentPlayer.CurrentCartridge >= CurrentPlayer.LiveCartridge {
		return PlayResult{false, nil, nil, fmt.Errorf("Invalid Turn: Player %d is dead.", CurrentPlayer.Id)}
	}
	if t.Action == Play {
		badCards := []Card{}
		newPlayerCards := make([]Card, len(CurrentPlayer.Cards))
		copy(newPlayerCards, CurrentPlayer.Cards)

		for _, c := range t.Cards {
			if slices.Contains(CurrentPlayer.Cards, c) {
				newPlayerCards = remove(newPlayerCards, c)
			} else {
				badCards = append(badCards, c)
			}
		}
		if len(badCards) > 0 {
			return PlayResult{false, nil, nil, errors.New("Played cards that didn't exist: " + CardListToString(badCards))}
		}
		CurrentPlayer.Cards = newPlayerCards
		if e.AllHandsEmpty() {
			return e.Play(Turn{Action: Challenge})
		} else {
			e.GameState.TurnHistory = append(e.GameState.TurnHistory, t)
			gameState.CardsLastPlayed = t.Cards
			gameState.PreviousPlayer = gameState.CurrentPlayer
			e.nextPlayer()
			return PlayResult{false, nil, nil, nil}
		}
	} else if t.Action == Challenge {
		cardsLastPlayed := gameState.CardsLastPlayed
		if len(gameState.CardsLastPlayed) <= 0 {
			return PlayResult{false, nil, nil, errors.New("Invalid challenge: No cards have been played yet")}
		}
		for _, c := range gameState.CardsLastPlayed {
			if c != Joker && c != gameState.TableCard {
				PreviousPlayer.CurrentCartridge++
				e.ResetRound()
				return PlayResult{false, cardsLastPlayed, PreviousPlayer, nil}
			}
		}

		CurrentPlayer.CurrentCartridge++
		e.ResetRound()
		return PlayResult{true, cardsLastPlayed, CurrentPlayer, nil}
	} else {
		return PlayResult{false, nil, nil, errors.New("Unknown move: " + strconv.Itoa(int(t.Action)))}
	}
}
