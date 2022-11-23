package runner

import (
	"log"
	"sync"

	"github.com/xenomote/sc2api/client"
)

// Run starts the game.
func Run(config *gameConfig) {
	config.StartAll()
	config.StartGame(mapPath())

	clients := config.clients

	wg := sync.WaitGroup{}
	wg.Add(len(clients))

	for _, c := range clients {
		go func(client *client.Client) {
			runAgent(client)
			cleanup(client)

			wg.Done()
		}(c)
	}

	wg.Wait()
}

func runAgent(c *client.Client) {
	defer func() {
		if p := recover(); p != nil {
			client.ReportPanic(p)
		}

		// If the bot crashed before losing, keep the game running (force the opponent to earn the win)
		for c.IsInGame() {
			if err := c.Step(224); err != nil { // 10 seconds per update
				log.Print(err)
				break
			}
		}
	}()

	// get GameInfo, Data, and Observation
	if err := c.Init(); err != nil {
		log.Printf("Failed to init client: %v", err)
		return
	}

	// make sure the bot was added to a game or replay
	if !c.IsInGame() {
		log.Print("Client is not in-game")
		return
	}

	// run the agent's code
	c.Agent.RunAgent(c)
}

func cleanup(c *client.Client) {
	c.RequestLeaveGame()

	// Print the winner
	for _, player := range c.Observation().GetPlayerResult() {
		if player.GetPlayerId() == c.PlayerID() {
			log.Print(player.GetResult())
		}
	}
}
