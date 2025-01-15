package main

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/FrankWhoee/ImposterInn/engine"
	"github.com/gofrs/uuid/v5"
	"github.com/olahol/melody"
)

// Pointer because maybe we want to pass it around in the future?
var idBroker *IdBroker

type MessageContext struct {
	cmd          string
	cmdargs      []string
	loginContext *LoginContext
	e            *engine.Engine
	m            *melody.Melody
}

type LoginContext struct {
	pid string
	wid int

	isInLobby bool
	lobbyId   string
}

func main() {
	idBroker = new(IdBroker)

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

	})

	m.HandleDisconnect(func(s *melody.Session) {
		if pid, ok := s.Get("id"); ok {
			removePlayer(pid.(string))
		}
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		smsg := strings.Trim(string(msg), " ")
		fmt.Println(smsg)

		cmd := smsg[0:4]
		cmdargs := strings.Split(strings.Trim(smsg[5:], "\n "), " ")
		var loginContext *LoginContext

		if pid, ok := s.Get("pid"); ok {
			loginContext = new(LoginContext)
			loginContext.pid = pid.(string)
			wid, isInLobby := idBroker.getWid(pid.(string))
			loginContext.isInLobby = isInLobby
			loginContext.wid = wid
		}

		messageContext := new(MessageContext)
		*messageContext = MessageContext{cmd, cmdargs, loginContext, e, m}

		msg_to_fn := make(map[string]func(*MessageContext))
		msg_to_fn["chal"] = chal
		// msg_to_fn["rqid"] =

		msg_to_fn[cmd](messageContext)

		if pid, ok := s.Get("pid"); ok && pidSet[pid.(string)] {
			spid := pid.(string)
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
				newpid := strings.Split(smsg, " ")[1]
				if b, ok := pidSet[newpid]; !ok || !b {
					wid := pidToWid[spid]
					widToPid[wid] = spid
					pidToWid[newpid] = wid
					pidSet[spid] = false
					pidSet[newpid] = true
					s.Write([]byte(fmt.Sprintf("assn %s %d", newpid, wid)))
				} else {
					rqid(m, s, e, bots)
				}
				return
			} else if smsg[0:4] == "prdy" {
				// TODO
			} else {
				return
			}
			continueUntilRealPlayer(m, e, bots)
			// sendHand(s, e.GameState.Players[pidToWid[spid]].Cards)
			privateBroadcastHand(m, e.GameState)
		} else if smsg[0:4] == "rqid" {
			rqid(m, s, e, bots)
		} else if smsg[0:4] == "iamp" {
			fmt.Println("iamp: could not find session")
			newpid := strings.Split(smsg, " ")[1]
			updateActiveConnections(m)
			if b, ok := pidSet[newpid]; !ok || !b {
				wid := findEmptyWid(m)
				pidSet[newpid] = true
				widToPid[wid] = newpid
				pidToWid[newpid] = wid
				s.Write([]byte(fmt.Sprintf("assn %s %d", newpid, wid)))
				continueUntilRealPlayer(m, e, bots)
				sendHand(s, e.GameState.Players[wid].Cards)
			} else {
				rqid(m, s, e, bots)
			}
			return
		}

	})

	http.ListenAndServe(":5000", nil)
}

func chal(mc *MessageContext) {
	t := engine.Turn{Action: engine.Challenge, Cards: nil, PlayerId: mc.loginContext.wid}
	cr, e := mc.e.Play(t)
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	broadcastChallengeResult(mc.m, cr)
}

func play(mc *MessageContext) {
	cardStrings := mc.cmdargs
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
	t := engine.Turn{Action: engine.Play, Cards: cards, PlayerId: mc.loginContext.wid}
	mc.e.Play(t)
}

func rqid(m *melody.Melody, s *melody.Session, e *engine.Engine, bots [4]engine.Bot) {
	pidbytes, _ := uuid.Must(uuid.NewV4()).MarshalText()
	pid := string(pidbytes)
	pidSet[pid] = true
	wid := findEmptyWid(m)

	widToPid[wid] = pid
	pidToWid[pid] = wid

	s.Set("pid", pid)
	continueUntilRealPlayer(m, e, bots)
	m.Broadcast([]byte(e.GameState.ToIIP()))
	sendHand(s, e.GameState.Players[wid].Cards)
	s.Write([]byte(fmt.Sprintf("assn %s %d", pid, wid)))
	fmt.Println(widToPid)
}

func findEmptyWid(m *melody.Melody) int {
	fmt.Println(widToPid)
	updateActiveConnections(m)
	wid := 0
	for i := 0; i < len(widToPid); i++ {
		wid = i
		if widToPid[i] == "" {
			break
		}
	}
	if widToPid[wid] != "" {
		wid++
	}
	return wid
}

// Go through all melody sessions and remove them if their session is closed
// impl. find active ids -> remove pid if not part of active ids
func updateActiveConnections(m *melody.Melody) {
	sessions, e := m.Sessions()
	if e != nil {
		return
	}
	active_ids := make([]string, 0)
	for _, s := range sessions {
		pid, ok := s.Get("pid")
		if ok && !s.IsClosed() {
			active_ids = append(active_ids, pid.(string))
		}
	}
	for pid := range pidSet {
		if !slices.Contains(active_ids, pid) {
			removePlayer(pid)
		}
	}
}

func removePlayer(pid string) {
	wid := pidToWid[pid]
	delete(pidSet, pid)
	delete(pidToWid, pid)
	delete(widToPid, wid)
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
