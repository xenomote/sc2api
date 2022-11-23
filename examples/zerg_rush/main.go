package main

import (
	"github.com/xenomote/sc2api/api"
	"github.com/xenomote/sc2api/client"
	"github.com/xenomote/sc2api/runner"
)

func main() {
	agent := client.AgentFunc(runAgent)
	
	// runner.SetRealtime()
	runner.Run(
		runner.NewGameConfig(
			client.NewParticipant(api.Race_Zerg, agent, "ZergRush"),
			// client.NewParticipant(api.Race_Zerg, agent, "ZergRush"),
			client.NewComputer(api.Race_Random, api.Difficulty_Medium, api.AIBuild_RandomBuild),
		),
	)
}
