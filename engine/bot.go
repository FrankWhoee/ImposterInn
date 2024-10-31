package engine

import (
	"math"
	"math/rand"
)

type Bot struct {
	Id int
}

func NewBot(id int) *Bot {
	b := new(Bot)
	b.Id = id
	return b
}

func floatN(minimum float64, maximum float64) float64{
	return minimum + rand.Float64() * (maximum - minimum)
}

func (b *Bot) NextMove(turnHistory []Turn, numCardsLastPlayed int, hand []Card, tableCard Card, prevPlayerTriggers int) Turn {
	numTableCards := 0
	for _, c := range hand {
		if c == Joker || c == tableCard {
			numTableCards++
		}
	}
	
	pOfTableCard := math.Pow((8 - float64(numTableCards))/15, float64(numCardsLastPlayed))
	pOfChallenge := min(1, max(0, 1 - pOfTableCard - (1/float64(6-prevPlayerTriggers)) + floatN(-0.1,0.1)))
	if len(turnHistory) > 0 && rand.Float64() < pOfChallenge {
		return Turn{Action: Challenge}
	} else {
		cardsToPlay := []Card{}
		if len(hand)-numTableCards > 0 && rand.Float32() > 0.5 {
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
}
