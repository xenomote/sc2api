package botutil

import (
	"github.com/chippydip/go-sc2ai/client"
)

// Bot ...
type Bot struct {
	client.AgentInfo

	*Player
	*UnitContext
	*Actions
	*Builder
}

// NewBot ...
func NewBot(info client.AgentInfo) Bot {
	bot := Bot{AgentInfo: info}

	bot.Player = NewPlayer(info)
	bot.Actions = NewActions(info)
	bot.UnitContext = NewUnitContext(info, bot.Actions)
	bot.Builder = NewBuilder(info, bot.Player, bot.UnitContext)

	return bot
}
