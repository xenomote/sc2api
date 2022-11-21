package runner

import (
	"math/rand"
	"path/filepath"
	"runtime"
	"time"
)

var (
	mapName = Random1v1Map()
)

func init() {
	flagStr("map", &mapName, "Which map to run.")
}

// SetMap sets the default map to use.
func SetMap(name string) {
	Set("map", name)
}

// Random1v1Map returns a random map name from the current 1v1 ladder map pool.
func Random1v1Map() string {
	currentMaps := maps2021season1

	rand.Seed(time.Now().UnixNano())
	return currentMaps[rand.Intn(len(currentMaps))] + ".SC2Map"
}

func mapPath() string {
	// Fix linux client using maps directory instead of Maps
	if runtime.GOOS != "windows" {
		return filepath.Join(defaultSc2Path(), "Maps", mapName)
	}
	return mapName
}

// TODO: check for current ladder pool maps, download if missing?

var Maps2018s3 = []string{
	"AcidPlantLE",
	"BlueshiftLE",
	"CeruleanFallLE",
	"DreamcatcherLE",
	"FractureLE",
	"LostAndFoundLE",
	"ParaSiteLE",
}

var Maps2018s4 = []string{
	"AutomatonLE",
	"BlueshiftLE",
	"CeruleanFallLE",
	"DarknessSanctuaryLE",
	"KairosJunctionLE",
	"ParaSiteLE",
	"PortAleksanderLE",
}

var Maps2019ladder8pre2 = []string{
	"Acropolis",
	"Artana",
	"CrystalCavern",
	"DigitalFrontier",
	"OldSunshine",
	"Treachery",
	"Triton",
}

var Maps2019ladder8 = []string{
	"AcropolisLE",
	"DiscoBloodbathLE",
	"EphemeronLE",
	"ThunderbirdLE",
	"TritonLE",
	"WintersGateLE",
	"WorldofSleepersLE",
}

var maps2021season1 = []string{
	"DeathAura506",
	"EternalEmpire506",
	"EverDream506",
	"GoldenWall506",
	"IceandChrome506",
	"PillarsofGold506",
	"Submarine506",
}
