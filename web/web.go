package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"github.com/olahol/melody"
	// "github.com/gofrs/uuid/v5"
	"github.com/FrankWhoee/ImposterInn/engine"
)

var idCounter atomic.Int64

// var u1 = uuid.Must(uuid.NewV4())

func main() {
	m := melody.New()
	e := engine.NewEngine()
	bots := [3]engine.Bot{engine.Bot{Id: 1}, engine.Bot{Id: 2}, engine.Bot{Id: 3}}
	continueUntilPlayer0(m, e, bots)
	fmt.Println(engine.CardListToString(e.GameState.CurrentPlayer().Cards))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/frontend/index.html")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	m.HandleConnect(func(s *melody.Session) {
		id := idCounter.Add(1)

		s.Set("id", id)
		m.Broadcast([]byte(e.GameState.ToIIP()))
		sendHand(s, e.GameState.Players[0].Cards)
		s.Write([]byte(fmt.Sprintf("iam %d", id)))
	})

	m.HandleDisconnect(func(s *melody.Session) {
		if id, ok := s.Get("id"); ok {
			m.BroadcastOthers([]byte(fmt.Sprintf("dis %d", id)), s)
		}
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		if _, ok := s.Get("id"); ok {

			smsg := string(msg)
			fmt.Println(smsg)
			if smsg[0:4] == "chal" {
				t := engine.Turn{Action: engine.Challenge, Cards: nil, PlayerId: 0}
				cr,e := e.Play(t)
				if e != nil {
					fmt.Println(e.Error())
					return
				}
				broadcastChallengeResult(m, cr)
			} else if smsg[0:4] == "play" {
				cardStrings := strings.Split(strings.Trim(smsg[5:], "\n "), " ")
				fmt.Println(cardStrings)
				cards := make([]engine.Card, 0)
				for _, cs := range cardStrings {
					cardint, e := strconv.Atoi(cs)
					card := engine.Card(cardint)
					if e == nil {
						cards = append(cards, card)
					}
				}
				fmt.Println(engine.CardListToString(cards))
				t := engine.Turn{Action: engine.Play, Cards: cards, PlayerId: 0}
				e.Play(t)
			} else {
				return
			}

			continueUntilPlayer0(m, e, bots)
			m.Broadcast([]byte(e.GameState.ToIIP()))
			sendHand(s, e.GameState.Players[0].Cards)
		}
	})

	http.ListenAndServe(":5000", nil)
}

func sendHand(s *melody.Session, cards []engine.Card) {
	var shand strings.Builder
	shand.WriteString("hand ")
	for _, c := range cards {
		shand.Write([]byte(fmt.Sprintf("%d ", c)))
	}
	s.Write([]byte(shand.String()))
}

func broadcastChallengeResult(m *melody.Melody, cr *engine.ChallengeResult) {
	m.Broadcast([]byte(cr.ToIIP()))
}

func continueUntilPlayer0(m *melody.Melody, e *engine.Engine, bots [3]engine.Bot) {
	for e.GameState.CurrentPlayerId != 0 {
		CurrentPlayer := e.GameState.CurrentPlayer()
		PreviousPlayer := e.GameState.PreviousPlayer()
		t := bots[e.GameState.CurrentPlayerId-1].NextMove(e.GameState.TurnHistory, len(e.GameState.CardsLastPlayed), CurrentPlayer.Cards, e.GameState.TableCard, PreviousPlayer.CurrentCartridge)
		if t.Action == engine.Challenge {
			fmt.Printf("Player %d challenges player %d.\n", CurrentPlayer.Id, PreviousPlayer.Id)
			cr,e := e.Play(t)
			broadcastChallengeResult(m, cr)
			if e != nil {
				fmt.Println(e.Error())
				continue
			}
		} else {
			fmt.Printf("Player %d claims %d %ss.\n", CurrentPlayer.Id, len(t.Cards), e.GameState.TableCard.String())
		}
		e.Play(t)
		m.Broadcast([]byte(e.GameState.ToIIP()))
	}
}
