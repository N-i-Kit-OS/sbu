package commands

import "fmt"

func printUsage() {
	fmt.Println("Usage: sbu <command> [config_path]")
	fmt.Println("Commands:")
	fmt.Println("  init [config_path]       create example config (default: config.yaml)")
	fmt.Println("  run  [config_path]       run backup or restore (default: config.yaml)")
}
