package runner

import (
	"log"
	"sync"

	"github.com/xenomote/sc2api/api"
	"github.com/xenomote/sc2api/client"
)

const launchPortStart = 8168

type gameConfig struct {
	netAddress string
	nextPort   int32

	processInfo []client.ProcessInfo
	playerSetup []*api.PlayerSetup
	ports       *client.Ports

	clients []*client.Client
	started bool
}

func NewGameConfig(participants ...client.PlayerSetup) *gameConfig {
	config := &gameConfig{
		netAddress: "127.0.0.1",
		nextPort:   launchPortStart,
	}

	for _, p := range participants {
		if p.Agent != nil {
			config.clients = append(config.clients, &client.Client{Agent: p.Agent})
		}
		config.playerSetup = append(config.playerSetup, p.PlayerSetup)
	}
	return config
}

func (config *gameConfig) StartGame(mapPath string) {
	if !config.createGame(mapPath) {
		log.Fatal("Failed to create game.")
	}

	clients := config.processInfo
	if len(clients) > 1 {
		server := clients[0]

		clientports := make([]*api.PortSet, len(clients)-1)
		for i := 1; i < len(clients); i++ {
			client := clients[i]

			clientports[i-1] = &api.PortSet{
				BasePort: client.BasePort,
				GamePort: client.GamePort,
			}
		}

		config.ports = &client.Ports{
			ServerPorts: api.PortSet{
				GamePort: server.GamePort,
				BasePort: server.BasePort,
			},

			ClientPorts: clientports,
		}

	}

	config.JoinGame()
}

func (config *gameConfig) createGame(mapPath string) bool {
	if !config.started {
		log.Panic("Game not started")
	}

	// Create with the first client
	err := config.clients[0].CreateGame(mapPath, config.playerSetup, processRealtime)
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}

func (config *gameConfig) JoinGame() bool {
	clients := config.clients
	var wg sync.WaitGroup
	wg.Add(len(clients))

	for i := range clients {
		go func(i int) {
			defer wg.Done()
			if err := clients[i].RequestJoinGame(config.playerSetup[i], processInterfaceOptions, config.ports); err != nil {
				log.Fatalf("Unable to join game: %v", err)
			}
		}(i)
	}

	wg.Wait()

	return true
}

func (config *gameConfig) Connect(port int32) {
	// Set process info for each bot
	config.processInfo = make([]client.ProcessInfo, len(config.clients))
	for i := range config.clients {
		config.processInfo[i] = client.ProcessInfo{GamePort: port}
	}

	// Since connect is blocking do it after the processes are launched.
	for i, client := range config.clients {
		pi := config.processInfo[i]

		if err := client.Connect(config.netAddress, pi.GamePort, processConnectTimeout); err != nil {
			log.Panic("Failed to connect")
		}
	}

	// Assume starcraft has started after succesfully attaching to a server
	config.started = true
}
