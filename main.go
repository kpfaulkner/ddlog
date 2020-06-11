package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kpfaulkner/ddlog/pkg"
	"github.com/kpfaulkner/ddlog/pkg/models"
	"log"
	"os"
	"os/user"
	"strings"
	"time"
)


// read config from multiple locations.
// first try local dir...
// if fails, try ~/.ddlog/config.json
func readConfig() models.Config{
	var configFile *os.File
	var err error
	configFile, err = os.Open("config.json")
	if err != nil {
		// try and read home dir location.
		usr, err := user.Current()
		if err != nil {
			log.Fatal( err )
		}
		configPath := fmt.Sprintf("%s/.ddlog/config.json", usr.HomeDir)
		configFile, err = os.Open(configPath)
		if err != nil {
			log.Fatal( err )
		}
	}
	defer configFile.Close()

	config := models.Config{}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}

func generateQuery( env string, level string, query string) string {
	queryTemplate := "@environment:%s status:(%s)"

	if strings.TrimSpace(query) != "" {
		queryTemplate = queryTemplate + ` "%s"`
		return fmt.Sprintf(queryTemplate, env,level,query)
	}

	return fmt.Sprintf(queryTemplate, env, level)
}

// just count for now... needs to add a lot more :)
func displayStats(resp *models.DatadogQueryResponse) {
	counts := len(resp.Logs)
	fmt.Printf("Result count %d\n", counts)
}

func displayResults(resp *models.DatadogQueryResponse) {
	for _,l := range resp.Logs {
		fmt.Printf("%s : %s\n", l.Content.Timestamp, l.Content.Message)
	}
}

func main() {
	fmt.Printf("So it begins...\n")

	env := flag.String("env", "prod", "environment: test,stage,prod")
	level := flag.String("level", "error", "level of logs to query against. info, warn, error")
	query := flag.String("query", "", "Part of the query that is NOT specifying level or env.")
	lastNMins := flag.Int("mins", 15, "Last N minutes to be searched")
	stats := flag.Bool("stats", false, "Give summary/stats of logs as opposed to raw logs.")

	flag.Parse()

	config := readConfig()
	dd := pkg.NewDatadog(config.DatadogAPIKey, config.DatadogAppKey)


	startDate := time.Now().UTC().Add( time.Duration(-1 * (*lastNMins)) * time.Minute)
	endDate := time.Now()

	formedQuery := generateQuery(*env, *level, *query)
	resp, err := dd.QueryDatadog(formedQuery, startDate, endDate)
	if err != nil {
		fmt.Printf("ERROR %s\n", err.Error())
		return
	}

	if *stats {
		// just the stats. :)
		displayStats(resp)

	} else {
		displayResults(resp)
	}
}
