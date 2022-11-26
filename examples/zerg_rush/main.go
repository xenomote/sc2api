package main

import (
	"flag"
	"log"
	"os"

	"github.com/xenomote/sc2api/api"
	"github.com/xenomote/sc2api/client"
	"github.com/xenomote/sc2api/runner"
)

func main() {
	f := flag.NewFlagSet("zerg rush", flag.ExitOnError)
	r := f.String("race", "", "the race you will play, defaults to random")
	s := f.Bool("fast", false, "whether the game should run fast")
	f.Parse(os.Args[1:])

	agent := client.AgentFunc(runAgent)
	ai := client.NewParticipant(api.Race_Zerg, agent, "ZergRush")

	var player, opponent client.PlayerSetup
	if *r == "" {
		if *s {
			runner.SetRealtime()
		}

		player = ai
		opponent = client.NewComputer(api.Race_Random, api.Difficulty_CheatInsane, api.AIBuild_Macro)
	} else {
		rv, ok := api.Race_value[*r]
		if !ok {
			log.Fatalln(*r, "is not a valid input for race")
		}
		race := api.Race(rv)

		runner.SetRealtime()
		human := client.AgentFunc(func(ai client.AgentInfo) {})
		player = client.NewParticipant(race, human, "Player")
		opponent = ai
	}
	
	runner.Run(runner.NewGameConfig(player, opponent))
}
