package client

import "github.com/xenomote/sc2api/api"

// PlayerSetup ...
type PlayerSetup struct {
	*api.PlayerSetup
	Agent
}

// NewParticipant ...
func NewParticipant(race api.Race, agent Agent, name string) PlayerSetup {
	return PlayerSetup{
		PlayerSetup: &api.PlayerSetup{
			Type:       api.PlayerType_Participant,
			Race:       race,
			PlayerName: name,
		},
		Agent: agent,
	}
}

// NewComputer ...
func NewComputer(race api.Race, difficulty api.Difficulty, build api.AIBuild) PlayerSetup {
	return PlayerSetup{
		PlayerSetup: &api.PlayerSetup{
			Type:       api.PlayerType_Computer,
			Race:       race,
			Difficulty: difficulty,
			AiBuild:    build,
		},
	}
}

// NewObserver ...
func NewObserver(agent Agent, name string) PlayerSetup {
	return PlayerSetup{
		PlayerSetup: &api.PlayerSetup{
			Type:       api.PlayerType_Observer,
			PlayerName: name,
		},
		Agent: agent,
	}
}