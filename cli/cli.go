package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/FrankWhoee/ImposterInn/engine"
)

func main() {
	e := engine.NewEngine()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("GAME START")
	fmt.Println(e.GameState.TableCard.String() + "'s Table")
	fmt.Printf("PLAYER %d's TURN - %d triggers pulled\n", e.GameState.CurrentPlayer, e.GameState.Players[e.GameState.CurrentPlayer].CurrentCartridge)
	fmt.Println("YOUR CARDS:")
	fmt.Println(engine.CardListToString(e.GameState.Players[e.GameState.CurrentPlayer].Cards))
	for scanner.Scan() {
		a := scanner.Text()
		println("-----------------")
		if strings.Trim(a, "\n ") == "Challenge" {
			playResult := e.Play(engine.Turn{Action: engine.Challenge, Cards: []engine.Card{}})
			if playResult.E != nil {
				fmt.Println(playResult.E.Error())
				continue
			}
			fmt.Println("CARDS REVEALED:")
			fmt.Println(engine.CardListToString(playResult.CardReveal))
			var survival string
			if playResult.TriggerPlayer.IsAlive() {
				survival = "survived"
			} else {
				survival = "died"
			}

			fmt.Printf("Player %d pulled the trigger. They %s on trigger %d.\n", playResult.TriggerPlayer.Id, survival, playResult.TriggerPlayer.CurrentCartridge)
		} else {
			cardStrings := strings.Split(strings.Trim(a, "\n "), " ")
			playedCards := []engine.Card{}
			var parseError error
			for _, s := range cardStrings {
				card, e := engine.StringToCard(strings.Trim(s, "\n"))
				parseError = e
				if e == nil {
					playedCards = append(playedCards, card)
				}
			}
			if parseError != nil {
				fmt.Println(parseError.Error())
				continue
			}
			playResult := e.Play(engine.Turn{Action: engine.Play, Cards: playedCards})
			if playResult.E != nil {
				fmt.Println(playResult.E.Error())
				continue
			}
		}
		if e.Winner() != nil {
			fmt.Printf("--------------PLAYER %d WINS--------------\n", e.Winner().Id)
			return
		}
		println("---------------------------------------------")
		fmt.Println(e.GameState.TableCard.String() + "'s Table")
		fmt.Printf("PLAYER %d's TURN - %d triggers pulled\n", e.GameState.CurrentPlayer, e.GameState.Players[e.GameState.CurrentPlayer].CurrentCartridge)
		fmt.Println("YOUR CARDS:")
		fmt.Println(engine.CardListToString(e.GameState.Players[e.GameState.CurrentPlayer].Cards))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
