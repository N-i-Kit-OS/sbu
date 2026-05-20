package commands

import (
	sbuserve "diplom/internal/sbuServe"
	"os"
)

func Execute() error {

	// check flags
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	// check command
	command := os.Args[1]

	// find config
	var configPath string
	if len(os.Args) > 2 {
		configPath = os.Args[2]
	} else {
		configPath = "config.yml"
	}

	switch command {

	case "init":
		return runInitCommand(configPath)

	case "run":

		// read config
		err := runRunCommand(configPath)
		if err != nil {
			return err
		}

	case "ui":
		err := sbuserve.StartServer()
		if err != nil {
			return err
		}

	case "sls":
		err := runSLS(configPath)
		if err != nil {
			return err
		}
	default:
		printUsage()
	}

	return nil
}
