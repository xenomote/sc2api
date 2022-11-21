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
	// Play a random map against a medium difficulty computer
	runner.SetComputer(api.Race_Random, api.Difficulty_Easy, api.AIBuild_RandomBuild)

	// Create the agent and then start the game
	botutil.SetGameVersion()
	agent := client.AgentFunc(runAgent)
	runner.RunAgent(client.NewParticipant(api.Race_Protoss, agent, "StubBot"))
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
