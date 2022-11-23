package runner

import (
	"log"
	"sync"

	"github.com/xenomote/sc2api/client"
)

var (
	ladderGamePort   = 0
	ladderStartPort  = 0
	ladderServer     = ""
	ladderOpponentID = ""
)

func init() {
	// Ladder Flags
	flagInt("GamePort", &ladderGamePort, "Port of client to connect to")
	flagInt("StartPort", &ladderStartPort, "Starting server port")
	flagStr("LadderServer", &ladderServer, "Ladder server address")
	flagStr("OpponentId", &ladderOpponentID, "Ladder ID of the opponent (for learning bots)")
}

// OpponentID returns the current ladder opponent ID or an empty string.
func OpponentID() string {
	return ladderOpponentID
}

// Run starts the game.
func Run(config *gameConfig) {
	if !loadSettings() {
		return
	}

	clients := config.clients

	if ladderGamePort > 0 {
		log.Print("Connecting to port ", ladderGamePort)
		config.Connect(int32(ladderGamePort))
		config.JoinGame()
		log.Print(" Successfully joined game")
	} else {
		config.StartAll()
		config.StartGame(mapPath())
	}

	wg := sync.WaitGroup{}
	wg.Add(len(clients))

	for _, c := range clients {
		go func(client *client.Client) {
			defer wg.Done()

			runAgent(client)
			cleanup(client)
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
	if ladderGamePort == 0 {
		// Leave the game (but only in non-ladder games)
		c.RequestLeaveGame()
	}

	// Print the winner
	for _, player := range c.Observation().GetPlayerResult() {
		if player.GetPlayerId() == c.PlayerID() {
			log.Print(player.GetResult())
		}
	}
}
