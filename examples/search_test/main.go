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
	agent := client.AgentFunc(runAgent)
	runner.Run(
		runner.NewGameConfig(
			client.NewParticipant(api.Race_Protoss, agent, "SearchTest"),
			client.NewComputer(api.Race_Random, api.Difficulty_Medium, api.AIBuild_RandomBuild),
		),
	)
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
