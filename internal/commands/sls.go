package commands

import (
	"diplom/internal/config"
	"diplom/internal/list"
	"fmt"
)

func runSLS(path string) error {

	// read config
	cfg, err := config.ReadConfigF(path)
	if err != nil {
		return err
	}

	// get snapshots
	listSnap, err := list.GetAllSnapshots(cfg)
	if err != nil {
		return err
	}

	// print
	for _, v := range listSnap {
		fmt.Println(v)
	}
	return nil
}
