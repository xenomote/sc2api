package runner

import (
	"log"

	"github.com/xenomote/sc2api/api"
	"github.com/xenomote/sc2api/client"
)

type gameConfig struct {
	netAddress  string
	processInfo []client.ProcessInfo
	playerSetup []*api.PlayerSetup
	ports       client.Ports

	clients  []*client.Client
	started  bool
	lastPort int
}

func NewGameConfig(participants ...client.PlayerSetup) *gameConfig {
	config := &gameConfig{
		"127.0.0.1",
		nil,
		nil,
		client.Ports{},
		nil,
		false,
		0,
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
	// TODO: Make this parallel and get rid of the WaitJoinGame method
	for i, client := range config.clients {
		if err := client.RequestJoinGame(config.playerSetup[i], processInterfaceOptions, config.ports); err != nil {
			log.Fatalf("Unable to join game: %v", err)
		}
	}

	return true
}

func (config *gameConfig) Connect(port int) {
	// Set process info for each bot
	config.processInfo = make([]client.ProcessInfo, len(config.clients))
	for i := range config.clients {
		config.processInfo[i] = client.ProcessInfo{Port: port}
	}

	// Since connect is blocking do it after the processes are launched.
	for i, client := range config.clients {
		pi := config.processInfo[i]

		if err := client.Connect(config.netAddress, pi.Port, processConnectTimeout); err != nil {
			log.Panic("Failed to connect")
		}
	}

	// Assume starcraft has started after succesfully attaching to a server
	config.started = true
}

func (config *gameConfig) setupPorts(numAgents int, startPort int, checkSingle bool) {
	humans := numAgents
	if checkSingle {
		humans = 0
		for _, p := range config.playerSetup {
			if p.Type == api.PlayerType_Participant {
				humans++
			}
		}
	}

	if humans > 1 {
		ports := config.ports
		ports.SharedPort = int32(startPort + 1)
		ports.ServerPorts = &api.PortSet{
			GamePort: int32(startPort + 2),
			BasePort: int32(startPort + 3),
		}

		for i := 0; i < numAgents; i++ {
			var base = int32(startPort + 4 + i*2)
			ports.ClientPorts = append(ports.ClientPorts, &api.PortSet{GamePort: base, BasePort: base + 1})
		}
		config.ports = ports
	}
}
