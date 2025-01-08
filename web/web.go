package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/FrankWhoee/ImposterInn/engine"
	"github.com/gofrs/uuid/v5"
	"github.com/olahol/melody"
)

var pidSet map[string]bool
var widToPid map[int]string
var pidToWid map[string]int

func main() {
	pidSet = make(map[string]bool)
	widToPid = make(map[int]string)
	pidToWid = make(map[string]int)

	m := melody.New()
	e := engine.NewEngine()
	bots := [4]engine.Bot{{Id: 0}, {Id: 1}, {Id: 2}, {Id: 3}}
	fmt.Println(engine.CardListToString(e.GameState.CurrentPlayer().Cards))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/frontend/index.html")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	m.HandleConnect(func(s *melody.Session) {
		pidbytes, _ := uuid.Must(uuid.NewV4()).MarshalText()
		pid := string(pidbytes)
		pidSet[pid] = true
		wid := 0
		for i := 0; i < len(widToPid); i++ {
			wid = i
			if widToPid[i] == "" {
				break
			}
		}
		if widToPid[wid] != "" {
			wid++
			widToPid[wid] = pid
			pidToWid[pid] = wid
		} else {
			widToPid[wid] = pid
			pidToWid[pid] = wid
		}

		s.Set("pid", pid)
		continueUntilLivePlayer(m, e, bots)
		m.Broadcast([]byte(e.GameState.ToIIP()))
		sendHand(s, e.GameState.Players[0].Cards)
		s.Write([]byte(fmt.Sprintf("assn %s", pid)))
	})

	m.HandleDisconnect(func(s *melody.Session) {
		if pid, ok := s.Get("id"); ok {
			m.BroadcastOthers([]byte(fmt.Sprintf("disc %d", pid)), s)
			wid := pidToWid[pid.(string)]
			pidSet[pid.(string)] = false
			delete(pidToWid, pid.(string))
			delete(widToPid, wid)
		}
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		if pid, ok := s.Get("pid"); ok && pidSet[pid.(string)] != false {
			smsg := strings.Trim(string(msg), " ")
			fmt.Println(smsg)
			if smsg[0:4] == "chal" {
				t := engine.Turn{Action: engine.Challenge, Cards: nil, PlayerId: 0}
				cr, e := e.Play(t)
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
			} else if smsg[0:4] == "iamp" {
				newid := strings.Split(smsg, " ")[1]
				if b, ok := pidSet[newid]; !ok || b == false {
					s.Write([]byte(fmt.Sprintf("assn %s", newid)))
					wid := pidToWid[pid.(string)]
					widToPid[wid] = pid.(string)
					pidToWid[newid] = wid
					pidSet[pid.(string)] = false
				} else {
					s.Write([]byte(fmt.Sprintf("assn %s", pid)))
				}
			} else {
				return
			}

			continueUntilLivePlayer(m, e, bots)
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

func continueUntilLivePlayer(m *melody.Melody, e *engine.Engine, bots [4]engine.Bot) {
	fmt.Println(widToPid)
	for _,ok := widToPid[e.GameState.CurrentPlayerId]; !ok; {
		fmt.Printf("%d %s %t", e.GameState.CurrentPlayerId, widToPid[e.GameState.CurrentPlayerId], ok)
		CurrentPlayer := e.GameState.CurrentPlayer()
		PreviousPlayer := e.GameState.PreviousPlayer()
		t := bots[e.GameState.CurrentPlayerId].NextMove(e.GameState.TurnHistory, len(e.GameState.CardsLastPlayed), CurrentPlayer.Cards, e.GameState.TableCard, PreviousPlayer.CurrentCartridge)
		if t.Action == engine.Challenge {
			fmt.Printf("Player %d challenges player %d.\n", CurrentPlayer.Id, PreviousPlayer.Id)
			cr, e := e.Play(t)
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
