package main

import (
	"log"

	"github.com/xenomote/sc2api/api"
	"github.com/xenomote/sc2api/botutil"
	"github.com/xenomote/sc2api/client"
	"github.com/xenomote/sc2api/enums/ability"
	"github.com/xenomote/sc2api/enums/buff"
	"github.com/xenomote/sc2api/enums/unit"
	"github.com/xenomote/sc2api/enums/upgrade"
	"github.com/xenomote/sc2api/enums/zerg"
	"github.com/xenomote/sc2api/search"
)

type bot struct {
	*botutil.Bot

	myStartLocation    api.Point2D
	myNaturalLocation  api.Point2D
	enemyStartLocation api.Point2D

	waveSize int

	hatcheryCount  int
	harvesterCount int

	harvestedGas int
}

func runAgent(info client.AgentInfo) {
	bot := bot{Bot: botutil.NewBot(info)}
	bot.LogActionErrors()

	bot.waveSize = 1
	bot.hatcheryCount = 1

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
	extractor := bot.Self[zerg.Extractor].First()

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

	if extractor.IsNil() {
		geyser := bot.Neutral.Vespene().ClosestTo(bot.myStartLocation)
		if !bot.BuildUnitOn(zerg.Drone, ability.Build_Extractor, geyser) {
			return
		}
	} else if !bot.HasUpgrade(upgrade.Zerglingmovementspeed) && bot.harvestedGas >= 100 {
		pool.Order(ability.Research_ZerglingMetabolicBoost)
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
	hatcheries := bot.Self[zerg.Hatchery]
	harvesters := bot.Self[zerg.Drone]

	if harvesters.Len() > bot.harvesterCount {
		bot.harvesterCount = harvesters.Len()

		for _, hatch := range hatcheries.Slice() {
			if hatch.IsBuilt() && hatch.GetAssignedHarvesters() >= hatch.GetIdealHarvesters() {
				continue
			}

			minerals := bot.Neutral.Minerals().ClosestTo(hatch.Pos2D())
			hatcheries.OrderTarget(ability.Rally_Workers, minerals)
		}
	}

	if hatcheries.Len() > bot.hatcheryCount {
		bot.hatcheryCount = hatcheries.Len()

		hatch := hatcheries.First()
		hatcheries.OrderPos(ability.Rally_Hatchery_Units, hatch.Pos2D())
	}

	extractor := bot.Self[zerg.Extractor].First()
	if !extractor.IsNil() {
		if bot.harvestedGas < 100 {
			bot.harvestedGas = int(bot.Vespene)
			if extractor.GetAssignedHarvesters() < extractor.GetIdealHarvesters() {
				harvesters.First().OrderTarget(ability.Harvest_Gather, extractor)
			}
		} else {
			if extractor.GetAssignedHarvesters() > 0 {
				minerals := bot.Neutral.Minerals().ClosestTo(extractor.Pos2D())
				harvesters.Choose(func(u botutil.Unit) bool {
					for _, order := range u.GetOrders() {
						if order.GetTargetUnitTag() == extractor.Tag {
							return true
						}
					}
					return false
				}).OrderTarget(ability.Harvest_Gather, minerals)
			}
		}
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
		bot.waveSize++
	}

	targets := bot.Enemy.Ground().All().Choose(func(u botutil.Unit) bool {
		switch u.UnitType {
		case unit.Zerg_Larva, unit.Zerg_Egg, unit.Zerg_Broodling:
			return false
		default:
			return true
		}
	})

	if targets.Len() == 0 {
		attacking.OrderPos(ability.Attack_Attack, bot.enemyStartLocation)
		return
	}

	attacking.Each(func(ling botutil.Unit) {
		target := targets.ClosestTo(ling.Pos2D())

		if ling.Pos2D().Distance2(target.Pos2D()) > 4*4 {
			// If target is far, attack it as unit, ling will run ignoring everything else
			ling.OrderTarget(ability.Attack_Attack, target)
		} else {
			// Attack as position, ling will choose best target around
			ling.OrderPos(ability.Attack_Attack, target.Pos2D())
		}
	})
}
