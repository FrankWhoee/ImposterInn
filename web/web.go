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
		continueUntilRealPlayer(m, e, bots)
		m.Broadcast([]byte(e.GameState.ToIIP()))
		sendHand(s, e.GameState.Players[wid].Cards)
		s.Write([]byte(fmt.Sprintf("assn %s %d", pid, wid)))
		fmt.Println(widToPid)
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
		if pid, ok := s.Get("pid"); ok && pidSet[pid.(string)] {
			spid := pid.(string)
			smsg := strings.Trim(string(msg), " ")
			fmt.Println(smsg)
			if smsg[0:4] == "chal" && pidToWid[spid] == e.GameState.CurrentPlayerId {
				t := engine.Turn{Action: engine.Challenge, Cards: nil, PlayerId: 0}
				cr, e := e.Play(t)
				if e != nil {
					fmt.Println(e.Error())
					return
				}
				broadcastChallengeResult(m, cr)
			} else if smsg[0:4] == "play" && pidToWid[spid] == e.GameState.CurrentPlayerId {
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
				t := engine.Turn{Action: engine.Play, Cards: cards, PlayerId: pidToWid[spid]}
				e.Play(t)
			} else if smsg[0:4] == "iamp" {
				newid := strings.Split(smsg, " ")[1]
				if b, ok := pidSet[newid]; !ok || !b {
					s.Write([]byte(fmt.Sprintf("assn %s", newid)))
					wid := pidToWid[spid]
					widToPid[wid] = spid
					pidToWid[newid] = wid
					pidSet[spid] = false
				} else {
					s.Write([]byte(fmt.Sprintf("assn %s", pid)))
				}
				return
			} else {
				return
			}
			continueUntilRealPlayer(m, e, bots)
			// sendHand(s, e.GameState.Players[pidToWid[spid]].Cards)
			privateBroadcastHand(m, e.GameState)
		}
	})

	http.ListenAndServe(":5000", nil)
}

func sendHand(s *melody.Session, cards []engine.Card) {
	var shand strings.Builder
	shand.WriteString("hand ")
	shand.WriteString(engine.CardlistToIIP(cards))
	s.Write([]byte(shand.String()))
}

func broadcastChallengeResult(m *melody.Melody, cr *engine.ChallengeResult) {
	m.Broadcast([]byte(cr.ToIIP()))
}

func privateBroadcastHand(m *melody.Melody, g *engine.GameState) {
	for _, p := range g.Players {
		cards := p.Cards
		var shand strings.Builder
		shand.WriteString("hand ")
		shand.WriteString(engine.CardlistToIIP(cards))
		m.BroadcastFilter([]byte(shand.String()), func(s *melody.Session) bool {
			pid, ok := s.Get("pid")
			return ok && pid == widToPid[p.Id]
		})
	}
}

// func makePidCheck(pid string) {
// 	return func
// }

func continueUntilRealPlayer(m *melody.Melody, e *engine.Engine, bots [4]engine.Bot) {

	for {
		// a,b :=
		// fmt.Printf("%d %s %t\n", e.GameState.CurrentPlayerId, widToPid[e.GameState.CurrentPlayerId], ok)
		if e.Winner() != nil {
			m.Broadcast([]byte(fmt.Sprintf("winp %d %d %d", e.Winner().Id, e.Winner().CurrentCartridge, e.Winner().LiveCartridge)))
			break
		}
		_, ok := widToPid[e.GameState.CurrentPlayerId]
		if ok {
			m.Broadcast([]byte(e.GameState.ToIIP()))
			break
		}
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
