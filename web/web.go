package main

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"

	// "strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/golang/protobuf/proto"

	"github.com/FrankWhoee/ImposterInn/engine"
	"github.com/FrankWhoee/ImposterInn/web/dto/transfer"
	"github.com/olahol/melody"
)

// Pointer because maybe we want to pass it around in the future?
var idBroker *IdBroker

type Lobby struct {
	id     int
	users  []User
	status LobbyStatus
}

type LobbyStatus int

const (
	Waiting LobbyStatus = iota
	InGame
)

var lobbies map[int]*Lobby

var users map[string]*User

type MessageContext struct {
	cmd          string
	cmdargs      []string
	loginContext *LoginContext
	s            *melody.Session
	m            *melody.Melody
}

type LoginContext struct {
	lobby *Lobby
	user  *User
}

type User struct {
	username     string
	webid        string
	enginePlayer *engine.Player
}

func main() {
	m := melody.New()
	e := engine.NewEngine()
	fmt.Println(engine.CardListToString(e.GameState.CurrentPlayer().Cards))

	cmdToFn := make(map[string]func(*MessageContext))
	cmdToFn["lbcr"] = lbcr
	cmdToFn["rqid"] = rqid
	cmdToFn["amid"] = amid
	cmdToFn["name"] = name

	lobbies = make(map[int]*Lobby)
	users = make(map[string]*User)

	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request /: %s\n", r.PathValue("path"))
		http.ServeFile(w, r, "web/frontend/dist/index.html")
	})

	http.HandleFunc("GET /assets/{path}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Asset Request /assets/{path}: %s\n", r.PathValue("path"))
		http.ServeFile(w, r, "web/frontend/dist/assets/"+r.PathValue("path"))
	})

	http.HandleFunc("GET /images/{path}/{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Asset Request /{path}: %s\n", r.PathValue("path"))
		http.ServeFile(w, r, "web/frontend/dist/"+r.PathValue("path"))
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("WebSocket Request /ws/\n")
		m.HandleRequest(w, r)
	})

	http.HandleFunc("POST /lbcr", func(w http.ResponseWriter, r *http.Request) {
		request := &transfer.LbcrDTO{}
		reqbody,_ := io.ReadAll(r.Body)
		proto.Unmarshal(reqbody, request)

		newLobbyId := 111111 + rand.IntN(888888)
		lobbies[newLobbyId] = new(Lobby)
		lobbies[newLobbyId].id = newLobbyId
		lobbies[newLobbyId].users = make([]User, 0)
		lobbies[newLobbyId].status = Waiting

		lobbies[newLobbyId].users = append(lobbies[newLobbyId].users, *users[request.Webid])

		response := &transfer.LbcrResponseDTO{
			Lobbyid: int32(newLobbyId),
		}
		w.Write([]byte(fmt.Sprintf("lbid %d", newLobbyId)))
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
		fmt.Printf("cmd is %s\n", cmd)
		// cmdargs := make([]string, 0)

		messageContext := new(MessageContext)
		messageContext.cmd = cmd
		if len(smsg) > 4 {
			messageContext.cmdargs = strings.Split(strings.Trim(smsg[5:], "\n "), " ")
		}
		messageContext.s = s
		messageContext.m = m

		userid, ok := s.Get("id")
		if ok {
			messageContext.loginContext = new(LoginContext)
			messageContext.loginContext.user = users[userid.(string)]
		}

		cmdToFn[cmd](messageContext)
	})

	http.ListenAndServe(":5000", nil)
}

// (l)o(b)by (cr)eate: Handle client creating a new lobby
func lbcr(mc *MessageContext) {
	newLobbyId := 111111 + rand.IntN(888888)
	lobbies[newLobbyId] = new(Lobby)
	lobbies[newLobbyId].id = newLobbyId
	lobbies[newLobbyId].users = make([]User, 0)
	lobbies[newLobbyId].status = Waiting

	lobbies[newLobbyId].users = append(lobbies[newLobbyId].users, *mc.loginContext.user)

	mc.s.Write([]byte(fmt.Sprintf("lbid %d", newLobbyId)))
}

// (l)o(b)by (j)o(i)n: Handle client joining a lobby
func lbjn(mc *MessageContext) {
	lobbyId, atoierr := strconv.Atoi(mc.cmdargs[0])
	if atoierr != nil {
		return
	}
	lobby, ok := lobbies[lobbyId]
	if !ok {
		return
	}
	lobby.users = append(lobby.users, *mc.loginContext.user)
}

// (r)e(q)uest (id): Handle client requesting an id
func rqid(mc *MessageContext) {
	idbytes, _ := uuid.Must(uuid.NewV4()).MarshalText()
	id := string(idbytes)

	for _, ok := users[id]; ok; {
		idbytes, _ = uuid.Must(uuid.NewV4()).MarshalText()
		id = string(idbytes)
	}

	asid(mc.s, id)
	mc.s.Set("id", id)
}

// (as)sign (id): Assign id to a client
func asid(s *melody.Session, id string) {
	s.Write([]byte(fmt.Sprintf("asid %s", id)))
}

// I (am) (id): Handle client declaring their id
func amid(mc *MessageContext) {
	id := mc.cmdargs[0]
	mc.s.Set("id", id)
}

// Handle client declaring their username
func name(mc *MessageContext) {
	username := mc.cmdargs[0]

	webid, ok := mc.s.Get("id")
	if !ok {
		// TODO: Handle error
	}
	user := new(User)
	user.username = username
	user.webid = webid.(string)
	users[webid.(string)] = user
	fmt.Println("User " + username + " has joined the server\n")
	fmt.Println("All users: \n")
	for id, user := range users {
		fmt.Printf("User ID: %s, Username: %s\n", id, user.username)
	}
}
