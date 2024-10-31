package engine

import (
	"math/rand"
)

type Bot struct {
	PrevPlayer
	Id int
}

func NewBot(id int) *Bot {
	b := new(Bot)
	b.Id = id
	return b
}

func (b *Bot) nextMove(numCardsPlayed int, numCardsLastPlayed int, hand []Card, tableCard Card, prevPlayerTriggers int) Turn {

	// b.numCardsPlayedSoFar + numCardsLastPlayed
	// challengeScore = min(b.numCardsPlayedSoFar + numCardsLastPlayed - prevPlayerTriggers, 0) + rand.Intn(2)
	// max = 3 - 0
	cardsToPlay := []Card{}
	numTableCards := 0
	for _, c := range hand {
		if c == Joker || c == tableCard {
			numTableCards++
		}
	}
	if numTableCards <= 0 || rand.Float32() > 0.5 {
		numCardsToPlay := 1 + rand.Intn(len(hand)-numTableCards)
		for _, c := range hand {
			if numCardsToPlay <= 0 {
				break
			}
			if c != Joker && c != tableCard && numCardsToPlay > 0 {
				cardsToPlay = append(cardsToPlay, c)
				numCardsToPlay--
			}
		}
	} else {
		numCardsToPlay := 1 + rand.Intn(numTableCards)
		for _, c := range hand {
			if numCardsToPlay <= 0 {
				break
			}
			if c == Joker || c == tableCard {
				cardsToPlay = append(cardsToPlay, c)
				numCardsToPlay--
			}
		}
	}
	return Turn{Action: Play, Cards: cardsToPlay}

}
