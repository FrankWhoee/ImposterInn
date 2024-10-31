package engine

import (
	"fmt"
	"slices"
)

type Card int

const (
	King Card = iota
	Queen
	Ace
	Joker
)

var cardStrings = []string{"King", "Queen", "Ace", "Joker"}

func (c Card) String() string {
	return cardStrings[c]
}

func CardListToString(cards []Card) string {
	outputString := ""
	for _, c := range cards {
		outputString += c.String() + ", "
	}
	return outputString
}

func StringToCard(s string) (Card, error) {
	i := slices.Index(cardStrings, s)
	if i < 0 {
		return -1, fmt.Errorf("Invalid card string %s. Must belong to %s", s, cardStrings)
	}
	return Card(i), nil
}


type Action int

const (
	Play Action = iota
	Challenge
)

type Turn struct {
	Action Action
	Cards  []Card
	PlayerId int
}

type Player struct {
	Cards            []Card
	CurrentCartridge int
	LiveCartridge    int
	Id               int
}

func (p *Player) IsAlive() bool {
	return p.CurrentCartridge < p.LiveCartridge
}

type GameState struct {
	Players         []*Player
	TableCard       Card
	CardsLastPlayed []Card
	CurrentPlayer   int
	PreviousPlayer  int
	TurnHistory []Turn
}

func NewGameState() *GameState {
	gameState := new(GameState)
	gameState.TurnHistory = []Turn{}

	for i := 0; i < 4; i++ {
		gameState.Players = append(gameState.Players, new(Player))
	}

	return gameState
}