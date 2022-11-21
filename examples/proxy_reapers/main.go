package main

import (
	"github.com/xenomote/sc2api/api"
	"github.com/xenomote/sc2api/botutil"
	"github.com/xenomote/sc2api/client"
	"github.com/xenomote/sc2api/runner"
)

func main() {
	// Play a random map against a VeryHard difficulty computer
	runner.SetComputer(api.Race_Random, api.Difficulty_VeryHard, api.AIBuild_RandomBuild)

	// Create the agent and then start the game
	botutil.SetGameVersion()
	agent := client.AgentFunc(runAgent)
	runner.RunAgent(client.NewParticipant(api.Race_Terran, agent, "ProxyReapers"))
}
