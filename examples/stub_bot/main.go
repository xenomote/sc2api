package main

import (
	"log"

	"github.com/xenomote/sc2api/api"
	"github.com/xenomote/sc2api/botutil"
	"github.com/xenomote/sc2api/client"
	"github.com/xenomote/sc2api/runner"
)

type bot struct {
	*botutil.Bot
}

func main() {
	agent := client.AgentFunc(runAgent)
	runner.Run(
		runner.NewGameConfig(
			client.NewParticipant(api.Race_Protoss, agent, "StubBot"),
			client.NewComputer(api.Race_Random, api.Difficulty_Medium, api.AIBuild_RandomBuild),
		),
	)
}

func runAgent(info client.AgentInfo) {
	bot := bot{Bot: botutil.NewBot(info)}
	bot.LogActionErrors()

	bot.init()
	for bot.IsInGame() {
		bot.doSmt()

		if err := bot.Step(1); err != nil {
			log.Print(err)
			break
		}
	}
}

func (bot *bot) init() {
	// Send a friendly hello
	bot.Chat("(glhf)")
}

func (bot *bot) doSmt() {

}
