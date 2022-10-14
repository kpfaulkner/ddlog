package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/kpfaulkner/ddlog/pkg/models"
	"log"
	"os"
	"os/user"
)

// read config from multiple locations.
// first try local dir...
// if fails, try ~/.ddlog/config.json
func ReadConfig() models.Config {
	var configFile *os.File
	var err error
	configFile, err = os.Open("config.json")
	if err != nil {
		// try and read home dir location.
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		configPath := fmt.Sprintf("%s/.ddlog/config.json", usr.HomeDir)
		configFile, err = os.Open(configPath)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer configFile.Close()

	config := models.Config{}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}
