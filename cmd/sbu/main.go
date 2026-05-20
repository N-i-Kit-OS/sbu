package main

import (
	"diplom/internal/commands"
	"fmt"
	"os"
)

func main() {

	if err := commands.Execute(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		fmt.Println(err)
	}
}
