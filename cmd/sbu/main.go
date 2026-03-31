package main

import (
	"diplom/internal/storage"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	Source      string `yaml: "source"`
	Bucket      string `yaml: "bucket"`
	Endpoint    string `yaml: "endpoint"`    //"localhost:9000"
	AccessKeyID string `yaml: "accessKeyId"` //"admin"
	SecretKey   string `yaml: "secretKey"`   //"admin123"
	UseSSL      string `yaml: "useSSL"`      //false
}

func main() {

	// create flags
	run := flag.Bool("run", false, "Running program")
	help := flag.Bool("help", false, "Help")

	flag.Parse()

	// check flags
	if flag.NFlag() == 0 || *help {

		fmt.Println("You need to run the program with the --run flag")

	} else if *run {

		// check config
		files, err := os.ReadDir(".")
		if err != nil {
			fmt.Println(err)
		}

		stateOfSearch := false

		for _, file := range files {

			if file.Name() == "config.yaml" {

				content, err := os.ReadFile(file.Name())
				if err != nil {
					fmt.Println(err)
				}

				var conf config

				err = yaml.Unmarshal(content, &conf)
				if err != nil {
					fmt.Println(err)
				}

				err = storage.BackupToS3(conf.Source)
				if err != nil {
					fmt.Println(err)
				}
				stateOfSearch = true
				break
			}
		}

		if !stateOfSearch {
			var cfg = config{
				Source:      "Path from directory to copy",
				Endpoint:    "Is your endpoint, example: localhost:9000",
				AccessKeyID: "Is AccessKeyID, example: admin",
				SecretKey:   "Is SecretKey, example: admin123",
				UseSSL:      "UseSSL, example: false",
				Bucket:      "Bucket name",
			}
			data, err := yaml.Marshal(&cfg)
			if err != nil {
				fmt.Println(err)
			}

			err = os.WriteFile("config.yaml", data, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
