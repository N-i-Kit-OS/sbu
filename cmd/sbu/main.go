package main

import (
	"diplom/internal/config"
	"diplom/internal/restore"
	"diplom/internal/storage"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	nameConfigFile = "config.yml"
)

func main() {

	// create flags
	run := flag.Bool("run", false, "Running program")
	help := flag.Bool("help", false, "Help")
	recov := flag.Bool("recov", false, "Restore backup")

	flag.Parse()

	// check flags
	if flag.NFlag() == 0 || *help {

		fmt.Println("You need to run the program with the --run flag")

	}

	conf, err := readConfigF()
	if err != nil {
		fmt.Println(err)
	}

	if *run {

		// upload directory from config to s3
		err = storage.UploadFile(conf)
		if err != nil {
			fmt.Println(err)
		}

	}

	if *recov {

		// restore backup
		err = restore.Restore(conf)
		if err != nil {
			fmt.Println(err)
		}

	}

}

func createCnfigExemple() error {

	var conf = config.Config{
		Source:         "Path from directory to copy",
		Endpoint:       "Is your endpoint, example: localhost:9000",
		AccessKeyID:    "Is AccessKeyID, example: admin",
		SecretKey:      "Is SecretKey, example: admin123",
		UseSSL:         false,
		Bucket:         "Bucket name: bucket",
		FromRecovery:   "Path to recovery: dirName/ or path/to/file.txt",
		DateRecovery:   "Date recovery, exemple: 2024-11-21_19-30-00",
		PathToRecovery: "Where to recovery, exemple: dirName/",
	}

	data, err := yaml.Marshal(&conf)
	if err != nil {
		return err
	}

	// create file config.yml
	err = os.WriteFile(nameConfigFile, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func readConfigF() (config.Config, error) {

	// find config
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {

		if file.Name() == nameConfigFile {

			// read config
			content, err := os.ReadFile(file.Name())
			if err != nil {
				return config.Config{}, err
			}

			// parse config to struct
			var conf config.Config

			err = yaml.Unmarshal(content, &conf)
			if err != nil {
				return config.Config{}, err
			}

			return conf, nil
		}
	}

	// if config not found
	err = createCnfigExemple()
	if err != nil {
		return config.Config{}, err
	}

	fmt.Println("Create config file")

	return config.Config{}, nil
}
