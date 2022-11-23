package client

import (
	"github.com/xenomote/sc2api/api"
)

// ProcessInfo ...
type ProcessInfo struct {
	PID      int
	GamePort int32
	BasePort int32
}

// Ports ...
type Ports struct {
	ServerPorts api.PortSet
	ClientPorts []*api.PortSet
}
