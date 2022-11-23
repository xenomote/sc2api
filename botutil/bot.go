package botutil

import (
	"fmt"
	"log"

	"github.com/xenomote/sc2api/client"
 )

// Bot ...
type Bot struct {
	client.AgentInfo
	GameLoop uint32

	*Player
	*UnitContext
	*Actions
	*Builder
}

// NewBot ...
func NewBot(info client.AgentInfo) *Bot {
	bot := &Bot{AgentInfo: info}

	bot.Player = NewPlayer(info)
	bot.Actions = NewActions(info)
	bot.UnitContext = NewUnitContext(info, bot)
	bot.Builder = NewBuilder(info, bot.Player, bot.UnitContext)

	update := func() {
		bot.GameLoop = bot.Observation().GetObservation().GetGameLoop()
		log.SetPrefix(fmt.Sprintf("[%v] ", bot.GameLoop))

		if bot.GameLoop == 224 {
			bot.checkVersion()
		}
	}
	update()
	bot.OnAfterStep(update)

	return bot
}

func (bot *Bot) checkVersion() {
	c, ok := bot.AgentInfo.(*client.Client)
	if !ok {
		log.Print("Skipping version check") // Should only happen when AgentInfo is mocked
		return
	}

	// It's critical that data versions match, however, or generated IDs may be wrong
	if c.DataBuild != DataBuild {
		log.Println("data build", c.DataBuild, "does not match", DataBuild)
	}

	if  c.DataVersion != DataVersion {
		log.Println("data version", c.DataVersion, "does not match", DataVersion)
	}
}