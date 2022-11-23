package runner

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/xenomote/sc2api/client"
)

var (
	launchBaseBuild        = uint32(0)
	launchDataVersion      = ""
	launchExtraCommandArgs = []string(nil)
)

// SetGameVersion specifies a specific base game and data version to use when launching.
func SetGameVersion(baseBuild uint32, dataVersion string) {
	launchBaseBuild = baseBuild
	launchDataVersion = dataVersion
}

// StartAll starts an instance of Starcraft II for each client
func (config *gameConfig) StartAll() {
	
	config.processInfo = config.launchProcesses()

	config.started = true
}

// KillAll stops all instances of Starcraft II associated with the game config
func (config *gameConfig) KillAll() {
	for _, pi := range config.processInfo {
		if proc, err := os.FindProcess(pi.PID); err == nil && proc != nil {
			proc.Kill()
		}
	}
	config.processInfo = nil
}

func (config *gameConfig) launchProcesses() []client.ProcessInfo {
	clients := config.clients

	// Make sure we have a valid executable path
	path := processPathForBuild(launchBaseBuild)
	if _, err := os.Stat(path); err != nil {
		if len(path) > 0 {
			log.Fatalf("%v does not exist on your filesystem.", path)
		} else {
			log.Fatalf("Executable path can't be found, try running the StarCraft II executable first.")
		}
	}

	info := make([]client.ProcessInfo, len(clients))

	// Start an sc2 process for each bot
	var wg sync.WaitGroup
	for i, c := range clients {
		wg.Add(1)
		go func(i int, c *client.Client) {
			defer wg.Done()

			info[i] = config.launchAndAttach(c, path)
		}(i, c)
	}
	wg.Wait()

	return info
}

func (config *gameConfig) launchAndAttach(c *client.Client, path string) client.ProcessInfo {
	pi := client.ProcessInfo{
		GamePort: config.NextPort(),
		BasePort: config.NextPort(),
	}

	// See if we can connect to an old instance before launching
	if err := c.TryConnect(config.netAddress, pi.GamePort); err == nil {
		return pi
	}

	args := []string{
		"-listen", config.netAddress,
		"-port", strconv.Itoa(int(pi.GamePort)),

		// DirectX will fail if multiple games try to launch in fullscreen mode. Force them into windowed mode.
		"-displayMode", "0", 
	}
	if len(launchDataVersion) > 0 {
		args = append(args, "-dataVersion", launchDataVersion)
	}
	args = append(args, launchExtraCommandArgs...)

	// TODO: window size and position

	pi.PID = startProcess(path, args)
	if pi.PID == 0 {
		log.Print("Unable to start sc2 executable with path: ", path)
	} else {
		log.Printf("Launched SC2 (%v), PID: %v", path, pi.PID)
	}

	// Attach
	if err := c.Connect(config.netAddress, pi.GamePort, processConnectTimeout); err != nil {
		log.Panic("Failed to connect")
	}

	return pi
}

func (config *gameConfig) NextPort() int32 {
	config.nextPort++
	return config.nextPort
}

func startProcess(path string, args []string) int {
	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set the working directory on windows
	if runtime.GOOS == "windows" {
		_, exe := filepath.Split(path)
		dir := sc2Path(path)
		if strings.Contains(exe, "_x64") {
			dir = filepath.Join(dir, "Support64")
		} else {
			dir = filepath.Join(dir, "Support")
		}
		cmd.Dir = dir
	}

	if err := cmd.Start(); err != nil {
		log.Print(err)
		return 0
	}

	return cmd.Process.Pid
}
