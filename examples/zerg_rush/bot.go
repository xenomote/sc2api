package main

import (
	"log"

	"github.com/xenomote/sc2api/api"
	"github.com/xenomote/sc2api/botutil"
	"github.com/xenomote/sc2api/client"
	"github.com/xenomote/sc2api/enums/ability"
	"github.com/xenomote/sc2api/enums/buff"
	"github.com/xenomote/sc2api/enums/zerg"
	"github.com/xenomote/sc2api/search"
)

type bot struct {
	*botutil.Bot

	myStartLocation    api.Point2D
	myNaturalLocation  api.Point2D
	enemyStartLocation api.Point2D

	waveSize      int
	hatcherycount int
}

func runAgent(info client.AgentInfo) {
	bot := bot{Bot: botutil.NewBot(info)}
	bot.LogActionErrors()

	bot.waveSize = 1
	bot.hatcherycount = 1

	bot.init()
	for bot.IsInGame() {
		bot.strategy()
		bot.tactics()

		if err := bot.Step(1); err != nil {
			log.Print(err)
			break
		}
	}
}

func (bot *bot) init() {
	// My hatchery is on start position
	bot.myStartLocation = bot.Self[zerg.Hatchery].First().Pos2D()
	bot.enemyStartLocation = *bot.GameInfo().GetStartRaw().GetStartLocations()[0]

	// Find natural location
	expansions := search.CalculateBaseLocations(bot.Bot, false)
	query := make([]*api.RequestQueryPathing, len(expansions))
	for i, exp := range expansions {
		pos := exp.Location
		query[i] = &api.RequestQueryPathing{
			Start: &api.RequestQueryPathing_StartPos{
				StartPos: &bot.myStartLocation,
			},
			EndPos: &pos,
		}
	}
	resp := bot.Query(api.RequestQuery{Pathing: query})
	best, minDist := -1, float32(256)
	for i, result := range resp.GetPathing() {
		if result.Distance < minDist && result.Distance > 5 {
			best, minDist = i, result.Distance
		}
	}
	bot.myNaturalLocation = expansions[best].Location

	// Send a friendly hello
	bot.Chat("(glhf)")
}

func (bot *bot) strategy() {
	// Do we have a pool? if not, try to build one
	pool := bot.Self[zerg.SpawningPool].First()
	if pool.IsNil() {
		pos := bot.myStartLocation.Offset(bot.enemyStartLocation, 5)
		if !bot.BuildUnitAt(zerg.Drone, ability.Build_SpawningPool, pos) {
			return // save up
		}
	}

	hatches := bot.Self.Count(zerg.Hatchery)

	// Build overlords as needed (want at least 3 spare supply per hatch)
	if bot.FoodLeft() <= 3*hatches && bot.Self.CountInProduction(zerg.Overlord) == 0 {
		if !bot.BuildUnit(zerg.Larva, ability.Train_Overlord) {
			return // save up
		}
	}

	// Any more than 14 drones will delay the first round of lings (waiting for larva)
	maxDrones := 14
	if hatches > 1 {
		maxDrones = hatches * 16
	}

	// Build drones to our cap
	droneCount := bot.Self.CountAll(zerg.Drone)
	bot.BuildUnits(zerg.Larva, ability.Train_Drone, maxDrones-droneCount)

	// We need a pool before trying to build lings or queens
	if pool.IsNil() || pool.BuildProgress < 1 {
		return
	}

	// Spend any extra larva on zerglings
	bot.BuildUnits(zerg.Larva, ability.Train_Zergling, 100)

	// Get a queen for every hatch if we still have minerals
	bot.BuildUnits(zerg.Hatchery, ability.Train_Queen, hatches-bot.Self.CountAll(zerg.Queen))

	// Expand to natural (mostly just for the larva, but might as well put it in the right spot)
	if hatches < 2 {
		bot.BuildUnitAt(zerg.Drone, ability.Build_Hatchery, bot.myNaturalLocation)
	}
}

func (bot *bot) tactics() {
	hatcheries := bot.Self.All().IsTownHall()

	if hatcheries.Len() > bot.hatcherycount {
		bot.hatcherycount = hatcheries.Len()

		building := hatcheries.Choose(func(u botutil.Unit) bool { return !u.IsBuilt() }).First()
		target := bot.Neutral.Minerals().ClosestTo(building.Pos2D())

		hatcheries.OrderTarget(ability.Rally_Workers, target)
		hatcheries.OrderPos(ability.Rally_Hatchery_Units, target.Pos2D())
	}

	if bot.OpponentRace == api.Race_Terran {
		bot.waveSize = 3
	}

	hatcheries.IsBuilt().NoBuff(buff.QueenSpawnLarvaTimer).Each(func(u botutil.Unit) {
		bot.Self[zerg.Queen].CanOrder(ability.Effect_InjectLarva).ClosestTo(u.Pos2D()).OrderTarget(ability.Effect_InjectLarva, u)
	})

	lings := bot.Self[zerg.Zergling]

	preparing, attacking := lings.Partition(func(u botutil.Unit) bool {
		for _, order := range u.Orders {
			if order.AbilityId == ability.Attack_Attack {
				return false
			}
		}
		return true
	})

	waiting, _ := preparing.Partition(func(u botutil.Unit) bool {
		return len(u.Orders) == 0
	})

	if waiting.Len() >= 6*bot.waveSize {
		waiting.OrderPos(ability.Attack_Attack, bot.enemyStartLocation)
	}

	targets := bot.getTargets()
	if targets.Len() == 0 {
		attacking.OrderPos(ability.Attack_Attack, bot.enemyStartLocation)
		return
	}

	attacking.Each(func(ling botutil.Unit) {
		target := targets.ClosestTo(ling.Pos2D())
		if ling.Pos2D().Distance2(target.Pos2D()) > 4*4 {
			// If target is far, attack it as unit, ling will run ignoring everything else
			ling.OrderTarget(ability.Attack_Attack, target)
		} else if target.UnitType == zerg.ChangelingZergling || target.UnitType == zerg.ChangelingZerglingWings {
			// Must specificially attack changelings, attack move is not enough
			ling.OrderTarget(ability.Attack_Attack, target)
		} else {
			// Attack as position, ling will choose best target around
			ling.OrderPos(ability.Attack_Attack, target.Pos2D())
		}
	})
}

// Get the current target list, prioritizing good targets over ok targets
func (bot *bot) getTargets() botutil.Units {
	// Prioritize things that can fight back
	if targets := bot.Enemy.Ground().CanAttack().All(); targets.Len() > 0 {
		return targets
	}

	// Otherwise just kill all the buildings
	return bot.Enemy.Ground().Structures().All()
}
