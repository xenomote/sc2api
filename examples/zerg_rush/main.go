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
	r := f.String("race", api.Race_Random.String(), "the race you will play, defaults to random")
	f.Parse(os.Args[1:])
	
	rv, ok := api.Race_value[*r]
	if !ok {
		log.Fatalln(*r, "is not a valid input for race")
	}
	race := api.Race(rv)

	agent := client.AgentFunc(runAgent)
	human := client.AgentFunc(func(ai client.AgentInfo) {})

	runner.SetRealtime()
	runner.Run(
		runner.NewGameConfig(
			client.NewParticipant(race, human, "ZergRush"),
			client.NewParticipant(api.Race_Zerg, agent, "ZergRush"),
		),
	)
}
