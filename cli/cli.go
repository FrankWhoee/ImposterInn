package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/FrankWhoee/ImposterInn/engine"
	"github.com/fatih/color"
)

func processPlayResult(playResult engine.PlayResult, g *engine.GameState) {
	if playResult.TriggerPlayer != nil && playResult.E == nil {
		fmt.Println("CARDS REVEALED:")
		fmt.Println(engine.CardListToString(playResult.CardReveal))
		var survival string
		if playResult.TriggerPlayer.IsAlive() {
			survival = "survived"
		} else {
			survival = "died"
		}

		fmt.Printf("Player %d pulled the trigger. They %s on trigger %d.\n", playResult.TriggerPlayer.Id, survival, playResult.TriggerPlayer.CurrentCartridge)
		color.Red("------------ROUND RESET------------")
		printStats(g)
	}
}

func printStats(g *engine.GameState) {
	for _, p := range g.Players {
		if p.IsAlive() {
			color.Green("[ALIVE] Player %d, %d triggers.", p.Id, p.CurrentCartridge)
		} else {
			color.Red("[DEAD]  Player %d", p.Id)
		}
	}
}

func main() {
	e := engine.NewEngine()
	bots := [3]engine.Bot{engine.Bot{Id: 1}, engine.Bot{Id: 2}, engine.Bot{Id: 3}}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("GAME START")
	println("---------------------------------------------")
	fmt.Println(e.GameState.TableCard.String() + "'s Table")
	if e.GameState.CurrentPlayer != 0 {
		fmt.Printf("PLAYER %d's TURN - %d triggers pulled\n", e.GameState.CurrentPlayer, e.GameState.Players[e.GameState.CurrentPlayer].CurrentCartridge)
		CurrentPlayer := e.GameState.Players[e.GameState.CurrentPlayer]
		PreviousPlayer := e.GameState.Players[e.GameState.PreviousPlayer]
		t := bots[e.GameState.CurrentPlayer-1].NextMove(e.GameState.TurnHistory, len(e.GameState.CardsLastPlayed), CurrentPlayer.Cards, e.GameState.TableCard, PreviousPlayer.CurrentCartridge)
		if t.Action == engine.Challenge {
			fmt.Printf("Player %d challenges player %d.\n", CurrentPlayer.Id, PreviousPlayer.Id)
			playResult := e.Play(t)
			if playResult.E != nil {
				fmt.Println(playResult.E.Error())
			}
			processPlayResult(playResult, e.GameState)
		} else {
			fmt.Printf("Player %d claims %d %ss.\n", CurrentPlayer.Id, len(t.Cards), e.GameState.TableCard.String())
			playResult := e.Play(t)
			processPlayResult(playResult, e.GameState)
		}
		println("---------------------------------------------")
	}
	for {
		fmt.Printf("PLAYER %d's TURN - %d triggers pulled\n", e.GameState.CurrentPlayer, e.GameState.Players[e.GameState.CurrentPlayer].CurrentCartridge)
		if e.GameState.CurrentPlayer == 0 {
			fmt.Println("YOUR CARDS:")
			fmt.Println(engine.CardListToString(e.GameState.Players[e.GameState.CurrentPlayer].Cards))
			print("INPUT (Challenge/CARDS): ")
			scanner.Scan()
			a := scanner.Text()
			println("-----------------")
			if strings.Trim(a, "\n ") == "Challenge" {
				playResult := e.Play(engine.Turn{Action: engine.Challenge, Cards: []engine.Card{}})
				if playResult.E != nil {
					fmt.Println(playResult.E.Error())
					continue
				}
				processPlayResult(playResult, e.GameState)
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
				playResult := e.Play(engine.Turn{Action: engine.Play, Cards: playedCards, PlayerId: e.GameState.CurrentPlayer})
				if playResult.E != nil {
					fmt.Println(playResult.E.Error())
					continue
				}
			}

		} else {
			CurrentPlayer := e.GameState.Players[e.GameState.CurrentPlayer]
			PreviousPlayer := e.GameState.Players[e.GameState.PreviousPlayer]
			t := bots[e.GameState.CurrentPlayer-1].NextMove(e.GameState.TurnHistory, len(e.GameState.CardsLastPlayed), CurrentPlayer.Cards, e.GameState.TableCard, PreviousPlayer.CurrentCartridge)
			time.Sleep(1 * time.Second)
			if t.Action == engine.Challenge {
				fmt.Printf("Player %d challenges player %d.\n", CurrentPlayer.Id, PreviousPlayer.Id)
				playResult := e.Play(t)
				if playResult.E != nil {
					fmt.Println(playResult.E.Error())
					continue
				}
				processPlayResult(playResult, e.GameState)
			} else {
				fmt.Printf("Player %d claims %d %ss.\n", CurrentPlayer.Id, len(t.Cards), e.GameState.TableCard.String())
				playResult := e.Play(t)
				processPlayResult(playResult, e.GameState)
			}
		}

		winner := e.Winner()
		if winner != nil {
			color.Green("------------------------------------------\n")
			color.Green("|             PLAYER %d WINS             |\n", winner.Id)
			color.Green("|               %d TRIGGERS              |\n", winner.CurrentCartridge)
			color.Green("|           %d AWAY FROM DEATH           |\n", winner.LiveCartridge-winner.CurrentCartridge)
			color.Green("------------------------------------------\n")
			return
		}
		println("---------------------------------------------")
		fmt.Println(e.GameState.TableCard.String() + "'s Table")

	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
