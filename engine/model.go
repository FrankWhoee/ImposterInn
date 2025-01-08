package engine

import (
	"fmt"
	"slices"
	"strings"
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

func CardlistToIIP(cards []Card) string {
	var shand strings.Builder
	for _, c := range cards {
		shand.Write([]byte(fmt.Sprintf("%d ", c)))
	}
	return shand.String()
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

func StringToCardList(s string) ([]Card, error) {
	cardStrings := strings.Split(strings.Trim(s, "\n "), " ")
	cardList := []Card{}
	for _, s := range cardStrings {
		card, e := StringToCard(strings.Trim(s, "\n"))
		if e == nil {
			cardList = append(cardList, card)
		} else {
			return nil, e
		}
	}
	return cardList, nil
}

type Action int

const (
	Play Action = iota
	Challenge
)

type Turn struct {
	Action     Action
	Cards      []Card
	PlayerId   int
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
	Players          []*Player
	TableCard        Card
	CardsLastPlayed  []Card
	CurrentPlayerId  int
	PreviousPlayerId int
	TurnHistory      []Turn
}

func (g *GameState) ToIIP() string {
	var sb strings.Builder

	sb.WriteString("gast start\n")
	sb.WriteString(fmt.Sprintf("tbcd %d\n", g.TableCard))
	sb.WriteString(fmt.Sprintf("nclp %d\n", len(g.CardsLastPlayed)))
	sb.WriteString(fmt.Sprintf("cpid %d\n", g.CurrentPlayerId))
	sb.WriteString(fmt.Sprintf("ppid %d\n", g.PreviousPlayerId))
	sb.WriteString(fmt.Sprintf("tuhi %d\n", len(g.TurnHistory)))
	sb.WriteString(fmt.Sprintf("plys %d\n", len(g.Players)))
	sb.WriteString(g.PlayersToString())
	sb.WriteString(g.TurnHistoryToString())
	sb.WriteString("gast end")
	return sb.String()
}

func (g *GameState) PlayersToString() string {
	var sb strings.Builder

	for _, p := range g.Players {
		// interp. Player {p.Id} has {p.CurrentCartridge}
		alivebit := 0
		if p.IsAlive() {
			alivebit = 1
		}
		sb.WriteString(fmt.Sprintf("plyr%d %d %d\n", p.Id, p.CurrentCartridge, alivebit))
	}

	return sb.String()
}

func (g *GameState) TurnHistoryToString() string {
	var sb strings.Builder

	for i, t := range g.TurnHistory {
		// interp. On turn {i}, player {t.playerId} claims {len(t.Cards)} table cards.
		sb.WriteString(fmt.Sprintf("turn%d %d %d\n", i, t.PlayerId, len(t.Cards)))
	}

	return sb.String()
}

func (g *GameState) CurrentPlayer() *Player {
	return g.Players[g.CurrentPlayerId]
}

func (g *GameState) PreviousPlayer() *Player {
	return g.Players[g.PreviousPlayerId]
}

func NewGameState() *GameState {
	gameState := new(GameState)
	gameState.TurnHistory = []Turn{}

	for i := 0; i < 4; i++ {
		gameState.Players = append(gameState.Players, new(Player))
	}

	return gameState
}
