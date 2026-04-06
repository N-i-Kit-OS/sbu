package main

import (
	"diplom/internal/config"
	"diplom/internal/storage"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {

	// create flags
	run := flag.Bool("run", false, "Running program")
	help := flag.Bool("help", false, "Help")

	flag.Parse()

	// check flags
	if flag.NFlag() == 0 || *help {

		fmt.Println("You need to run the program with the --run flag")

	} else if *run {

		// find config
		files, err := os.ReadDir(".")
		if err != nil {
			fmt.Println(err)
		}

		stateOfSearch := false

		for _, file := range files {

			if file.Name() == "config.yml" {

				content, err := os.ReadFile(file.Name())
				if err != nil {
					fmt.Println(err)
				}

				var conf config.Config

				err = yaml.Unmarshal(content, &conf)
				if err != nil {
					fmt.Println(err)
				}

				err = storage.UploadFile(conf)
				if err != nil {
					fmt.Println(err)
				}
				stateOfSearch = true
				break
			}
		}

		if !stateOfSearch {
			var conf = config.Config{
				Source:      "Path from directory to copy",
				Endpoint:    "Is your endpoint, example: localhost:9000",
				AccessKeyID: "Is AccessKeyID, example: admin",
				SecretKey:   "Is SecretKey, example: admin123",
				UseSSL:      false,
				Bucket:      "Bucket name",
			}
			data, err := yaml.Marshal(&conf)
			if err != nil {
				fmt.Println(err)
			}

			err = os.WriteFile("config.yml", data, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
