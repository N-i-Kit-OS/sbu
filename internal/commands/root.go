package commands

import (
	"fmt"
	"os"
)

func Execute() error {
	if len(os.Args) < 2 {
		printUsage()
		return fmt.Errorf("no command")
	}

	command := os.Args[1]

	var configPath string
	if len(os.Args) > 2 {
		configPath = os.Args[2]
	} else {
		configPath = "config.yml"

	}

	switch command {
	case "init":
		return handleInit(configPath)
	case "run":
		return handleRun(configPath)
	case "ui":
		return handleUI()
	case "sls":
		return handleSLS(configPath)
	default:
		printUsage()
		return nil
	}
}
