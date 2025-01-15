package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/FrankWhoee/ImposterInn/engine"
	"github.com/olahol/melody"
)

// Pointer because maybe we want to pass it around in the future?
var idBroker *IdBroker

type MessageContext struct {
	cmd          string
	cmdargs      []string
	loginContext *LoginContext
	e            *engine.Engine
	s            *melody.Session
	m            *melody.Melody
}

type LoginContext struct {
	pid string
	wid int

	isInLobby bool
	lobbyId   string
}

func main() {
	idBroker = NewIdBroker()

	m := melody.New()
	e := engine.NewEngine()
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
		// if pid, ok := s.Get("id"); ok {
		// 	removePlayer(pid.(string))
		// }
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		smsg := strings.Trim(string(msg), " ")
		fmt.Println(smsg)

		cmd := smsg[0:4]
		cmdargs := make([]string, 0)
		if len(smsg) > 4 {
			cmdargs = strings.Split(strings.Trim(smsg[5:], "\n "), " ")
		}

		var loginContext *LoginContext

		if pid, ok := s.Get("pid"); ok {
			loginContext = new(LoginContext)
			loginContext.pid = pid.(string)
			wid, isInLobby := idBroker.getWid(pid.(string))
			loginContext.isInLobby = isInLobby
			loginContext.wid = wid
		}

		messageContext := new(MessageContext)
		*messageContext = MessageContext{cmd, cmdargs, loginContext, e, s, m}

		msg_to_fn := make(map[string]func(*MessageContext))
		msg_to_fn["chal"] = chal
		msg_to_fn["rqid"] = rqid
		msg_to_fn["play"] = play
		msg_to_fn["iamp"] = iamp

		msg_to_fn[cmd](messageContext)
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
	continueUntilRealPlayer(mc.m, mc.e)
	privateBroadcastHand(mc.m, mc.e.GameState)
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
	continueUntilRealPlayer(mc.m, mc.e)
}

func iamp(mc *MessageContext) {
	requestedPid := mc.cmdargs[0]
	if idBroker.isRegistered(requestedPid) {
		if wid, isInLobby := idBroker.getWid(requestedPid); isInLobby {
			mc.s.Write([]byte(fmt.Sprintf("assn %s %d", requestedPid, wid)))
			sendHand(mc.s, mc.e.GameState.Players[wid].Cards)
		} else {
			wid := idBroker.assignWid(requestedPid)
			mc.s.Write([]byte(fmt.Sprintf("assn %s %d", requestedPid, wid)))

			sendHand(mc.s, mc.e.GameState.Players[wid].Cards)
		}
		mc.s.Set("pid", requestedPid)
	} else {
		pid := idBroker.issuePid()
		wid := idBroker.assignWid(pid)
		mc.s.Write([]byte(fmt.Sprintf("assn %s %d", pid, wid)))
		mc.s.Set("pid", pid)
		sendHand(mc.s, mc.e.GameState.Players[wid].Cards)
	}
	continueUntilRealPlayer(mc.m, mc.e)
	mc.m.Broadcast([]byte(mc.e.GameState.ToIIP()))
}

func rqid(mc *MessageContext) {
	pid := idBroker.issuePid()
	wid := idBroker.assignWid(pid)
	mc.s.Write([]byte(fmt.Sprintf("assn %s %d", pid, wid)))
	mc.s.Set("pid", pid)

	continueUntilRealPlayer(mc.m, mc.e)
	mc.m.Broadcast([]byte(mc.e.GameState.ToIIP()))
	sendHand(mc.s, mc.e.GameState.Players[wid].Cards)
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
			sessionPid, isSession := s.Get("pid")
			brokerPid, isInLobby := idBroker.getPid(p.Id)
			return isSession && isInLobby && brokerPid == sessionPid
		})
	}
}

// func makePidCheck(pid string) {
// 	return func
// }

func continueUntilRealPlayer(m *melody.Melody, e *engine.Engine) {

	for {
		// a,b :=
		// fmt.Printf("%d %s %t\n", e.GameState.CurrentPlayerId, widToPid[e.GameState.CurrentPlayerId], ok)
		if e.Winner() != nil {
			m.Broadcast([]byte(fmt.Sprintf("winp %d %d %d", e.Winner().Id, e.Winner().CurrentCartridge, e.Winner().LiveCartridge)))
			break
		}
		_, ok := idBroker.getPid(e.GameState.CurrentPlayerId)
		if ok {
			m.Broadcast([]byte(e.GameState.ToIIP()))
			break
		}
		CurrentPlayer := e.GameState.CurrentPlayer()
		PreviousPlayer := e.GameState.PreviousPlayer()
		bot := engine.Bot{e.GameState.CurrentPlayerId}
		t := bot.NextMove(e.GameState.TurnHistory, len(e.GameState.CardsLastPlayed), CurrentPlayer.Cards, e.GameState.TableCard, PreviousPlayer.CurrentCartridge)
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
