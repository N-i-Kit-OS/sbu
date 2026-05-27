package commands

import "diplom/internal/config"

func handleInit(path string) error {
	return config.Init(path)
}
