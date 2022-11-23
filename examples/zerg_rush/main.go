package main

import (
	"flag"
	"os"

	"github.com/xenomote/sc2api/api"
	"github.com/xenomote/sc2api/client"
	"github.com/xenomote/sc2api/runner"
)

func main() {
	f := flag.NewFlagSet("zerg rush", flag.ExitOnError)
	r := f.String("race", api.Race_Terran.String(), "the race you will play")
	f.Parse(os.Args[1:])

	race := api.Race(api.Race_value[*r])

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
