package main

import (
	"fmt"
	"net/http"
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
	var gamelog strings.Builder
	continueUntilPlayer0(e, bots, &gamelog)
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
		m.Broadcast([]byte(fmt.Sprintf("gamestate ")))
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
				e.Play(t)
			} else if smsg[0:4] == "play" {
				cards, _ := engine.StringToCardList(smsg[5:])
				t := engine.Turn{Action: engine.Play, Cards: cards, PlayerId: 0}
				e.Play(t)
			} else {
				// what the fuck
			}
			fmt.Println(gamelog.String())
			fmt.Println(engine.CardListToString(e.GameState.CurrentPlayer().Cards))
			m.Broadcast([]byte(gamelog.String()))
			continueUntilPlayer0(e, bots, &gamelog)

			// m.BroadcastOthers([]byte(fmt.Sprintf("set %d %s", id, msg)), s)
		}
	})

	http.ListenAndServe(":5000", nil)
}

func continueUntilPlayer0(e *engine.Engine, bots [3]engine.Bot, gamelog *strings.Builder) {
	for e.GameState.CurrentPlayerId != 0 {
		CurrentPlayer := e.GameState.CurrentPlayer()
		PreviousPlayer := e.GameState.PreviousPlayer()
		t := bots[e.GameState.CurrentPlayerId-1].NextMove(e.GameState.TurnHistory, len(e.GameState.CardsLastPlayed), CurrentPlayer.Cards, e.GameState.TableCard, PreviousPlayer.CurrentCartridge)
		if t.Action == engine.Challenge {
			gamelog.WriteString(fmt.Sprintf("Player %d challenges player %d.\n", CurrentPlayer.Id, PreviousPlayer.Id))
			playResult := e.Play(t)
			if playResult.E != nil {
				fmt.Println(playResult.E.Error())
				continue
			}
		} else {
			gamelog.WriteString(fmt.Sprintf("Player %d claims %d %ss.\n", CurrentPlayer.Id, len(t.Cards), e.GameState.TableCard.String()))
		}
		e.Play(t)
	}
}
