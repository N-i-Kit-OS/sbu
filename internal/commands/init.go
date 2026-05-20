package commands

import "diplom/internal/config"

func runInitCommand(path string) error {
	return config.CreateExampleConfig(path)
}
