package main

import (
	"log"
	"time"

	"github.com/xenomote/sc2api/api"
	"github.com/xenomote/sc2api/botutil"
	"github.com/xenomote/sc2api/client"
	"github.com/xenomote/sc2api/runner"
	"github.com/xenomote/sc2api/search"
)

type bot struct {
	*botutil.Bot
}

func main() {
	// Play a random map against an easy difficulty computer
	runner.SetComputer(api.Race_Random, api.Difficulty_Easy, api.AIBuild_RandomBuild)

	// Create the agent and then start the game
	botutil.SetGameVersion()
	agent := client.AgentFunc(runAgent)
	runner.RunAgent(client.NewParticipant(api.Race_Protoss, agent, "SearchTest"))
}

func runAgent(info client.AgentInfo) {
	bot := bot{Bot: botutil.NewBot(info)}
	bot.LogActionErrors()
	bot.SetPerfInterval(224)

	search.CalculateBaseLocations(bot.Bot, true)
	pg := search.NewPlacementGrid(bot.Bot)

	for bot.IsInGame() {
		pg.DebugPrint(bot.Bot)

		if err := bot.Step(1); err != nil {
			log.Print(err)
			break
		}

		if bot.GameLoop > 224*30 {
			break
		}
	}

	for {
		time.Sleep(time.Second)
	}
}
