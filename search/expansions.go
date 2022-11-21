package search

import (
	"github.com/chippydip/go-sc2ai/api"
	"github.com/chippydip/go-sc2ai/botutil"
	"github.com/chippydip/go-sc2ai/client"
	"github.com/chippydip/go-sc2ai/enums/neutral"
)

// BaseLocation contains the optimal build location for a given resource cluster.
type BaseLocation struct {
	Resources UnitCluster
	Location  api.Point2D
}

// CalculateBaseLocations groups resources into clusters and determines the best town hall location for each cluster.
// The Center() point of each cluster is the optimal town hall location. If debug is true then the results will also
// be visualized in-game (until new debug info is drawn).
func CalculateBaseLocations(bot *botutil.Bot, debug bool) []BaseLocation {
	// Start by finding resource clusters
	clusters := Cluster(bot.Neutral.Resources().Choose(func(u botutil.Unit) bool {
		return u.UnitType != neutral.MineralField450 // mineral walls
	}), 15)

	// Add resource-restrictions to the placement grid
	placement := bot.GameInfo().StartRaw.PlacementGrid.Bits().ToBytes()
	bot.Neutral.Minerals().Each(func(u botutil.Unit) {
		markUnbuildable(placement, int32(u.Pos.X-0.5), int32(u.Pos.Y), 2, 1)
	})
	bot.Neutral.Vespene().Each(func(u botutil.Unit) {
		markUnbuildable(placement, int32(u.Pos.X-1), int32(u.Pos.Y-1), 3, 3)
	})

	// Mark locations which *can't* have town hall centers
	for y := int32(0); y < placement.Height(); y++ {
		for x := int32(0); x < placement.Width(); x++ {
			if placement.Get(x, y) < 128 {
				expandUnbuildable(placement, x, y)
			}
		}
	}

	// Find the nearest remaining square to each cluster's CoM
	locs := make([]BaseLocation, len(clusters))
	for i, cluster := range clusters {
		pt := cluster.Center()
		px, py := int32(pt.X), int32(pt.Y)
		var r2Min, xBest, yBest int32 = 256, -1, -1
		for r := int32(0); r*r <= r2Min; r++ { // search radius
			xMin, xMax, yMin, yMax := px-r, px+r, py-r, py+r
			for y := yMin; y <= yMax; y++ {
				for x := xMin; x <= xMax; x++ {
					// This is slightly inefficient, but much easier than repeating the same loop 4x for the edges
					if (x == xMin || x == xMax || y == yMin || y == yMax) && placement.Get(x, y) == 255 {
						dx, dy := x-px, y-py
						if r2 := dx*dx + dy*dy; r2 < r2Min {
							r2Min, xBest, yBest = r2, x, y
						}
					}
				}
			}
		}

		// Update the Center to be the detected location rather than the actual CoM (just don't add new units)
		locs[i] = BaseLocation{clusters[i], api.Point2D{X: float32(xBest) + 0.5, Y: float32(yBest) + 0.5}}
	}

	if debug {
		debugPrintBaseLocs(locs, placement, bot)
	}

	return locs
}

// markUnbuildable marks a w x h area around px, py (minus corners) as unbuildable (red)
func markUnbuildable(placement api.ImageDataBytes, px, py, w, h int32) {
	xMin, xMax := px-3, px+w+2
	yMin, yMax := py-3, py+h+2

	for y := yMin; y <= yMax; y++ {
		for x := xMin; x <= xMax; x++ {
			if (y == yMin || y == yMax) && (x == xMin || x == xMax) {
				continue // skip corners
			}
			if placement.Get(x, y) == 255 {
				placement.Set(x, y, 1)
			}
		}
	}
}

// expandUnbuildable marks any tile within 2 units of px, py as unbuildable (blue)
func expandUnbuildable(placement api.ImageDataBytes, px, py int32) {
	xMin, xMax := px-2, px+2
	yMin, yMax := py-2, py+2

	for y := yMin; y <= yMax; y++ {
		for x := xMin; x <= xMax; x++ {
			if placement.Get(x, y) == 255 {
				placement.Set(x, y, 128)
			}
		}
	}
}

// debugPrintBaseLocs shows debug info about the base location search procedure in-game
func debugPrintBaseLocs(locs []BaseLocation, placement api.ImageDataBytes, bot client.AgentInfo) {
	info := bot.GameInfo()
	heightMap := NewHeightMap(info.StartRaw)
	pathable := info.StartRaw.PathingGrid.Bits()

	var boxes []*api.DebugBox

	// Debug placement grid
	for y := int32(0); y < placement.Height(); y++ {
		for x := int32(0); x < placement.Width(); x++ {
			color := baseLocColor(placement.Get(x, y), pathable.Get(x, y))
			if color != nil {
				z := heightMap.Interpolate(float32(x)+0.5, float32(y)+0.5)
				boxes = append(boxes, &api.DebugBox{
					Color: color,
					Min:   &api.Point{X: float32(x) + 0.25, Y: float32(y) + 0.25, Z: z},
					Max:   &api.Point{X: float32(x) + 0.75, Y: float32(y) + 0.75, Z: z},
				})
			}
		}
	}

	// Base locations
	for _, exp := range locs {
		units := exp.Resources.Units()
		min, max := units[0].Pos2D(), units[0].Pos2D()
		for i := 1; i < len(units); i++ {
			p := units[i].Pos2D()
			if p.X < min.X {
				min.X = p.X
			}
			if p.X > max.X {
				max.X = p.X
			}
			if p.Y < min.Y {
				min.Y = p.Y
			}
			if p.Y > max.Y {
				max.Y = p.Y
			}
		}

		pt := exp.Location
		z := heightMap.Interpolate(pt.X+0.5, pt.Y+0.5)
		cm := exp.Resources.Center()
		cmz := heightMap.Interpolate(cm.X, cm.Y)
		boxes = append(boxes, &api.DebugBox{
			Color: green,
			Min:   &api.Point{X: pt.X - 2.5, Y: pt.Y - 2.5, Z: z},
			Max:   &api.Point{X: pt.X + 2.5, Y: pt.Y + 2.5, Z: z},
		}, &api.DebugBox{
			Color: green,
			Min:   &api.Point{X: pt.X - 0.05, Y: pt.Y - 0.05, Z: z},
			Max:   &api.Point{X: pt.X + 0.05, Y: pt.Y + 0.05, Z: z},
		}, &api.DebugBox{
			Color: white,
			Min:   &api.Point{X: cm.X - 0.05, Y: cm.Y - 0.05, Z: cmz - 1},
			Max:   &api.Point{X: cm.X + 0.05, Y: cm.Y + 0.05, Z: cmz + 1},
		}, &api.DebugBox{
			Color: white,
			Min:   &api.Point{X: min.X, Y: min.Y, Z: cmz - 1},
			Max:   &api.Point{X: max.X, Y: max.Y, Z: cmz + 1},
		})

	}

	bot.SendDebugCommands([]*api.DebugCommand{
		{
			Command: &api.DebugCommand_Draw{
				Draw: &api.DebugDraw{
					Boxes: boxes,
				},
			},
		},
		{
			Command: &api.DebugCommand_GameState{
				GameState: api.DebugGameState_show_map,
			},
		},
	})
}

// Re-use these colors so we don't have to keep allocating them
var (
	white = &api.Color{R: 255, G: 255, B: 255}
	red   = &api.Color{R: 255, G: 1, B: 1}
	blue  = &api.Color{R: 1, G: 1, B: 255}
	green = &api.Color{R: 1, G: 255, B: 1}
)

// baseLocColor converts a building placement value into a display color
func baseLocColor(value byte, pathable bool) *api.Color {
	switch value {
	case 255:
		return white // center buildable
	case 128:
		return blue // too close for center
	case 1:
		return red // too close to resources
	}
	if pathable {
		return green // not buildable, but pathable
	}
	return nil
}
