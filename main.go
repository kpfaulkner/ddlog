package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kpfaulkner/ddlog/pkg"
	"github.com/kpfaulkner/ddlog/pkg/models"
	"log"
	"os"
	"strings"
	"time"
)


func readConfig() models.Config{
	configFile, err := os.Open("config.json-real")
	defer configFile.Close()
	if err != nil {
		log.Panic("Unable to read config.json")
	}

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

func main() {
	fmt.Printf("So it begins...\n")

	env := flag.String("env", "prod", "environment: test,stage,prod")
	level := flag.String("level", "error", "level of logs to query against. info, warn, error")
	query := flag.String("query", "", "Part of the query that is NOT specifying level or env.")
	lastNMins := flag.Int("mins", 15, "Last N minutes to be searched")

	flag.Parse()

	config := readConfig()
	dd := pkg.NewDatadog(config.DatadogAPIKey, config.DatadogAppKey)


	// example query. "@environment:prod status:(error)", request.Range.From, request.Range.To

	startDate := time.Now().UTC().Add( time.Duration(-1 * (*lastNMins)) * time.Minute)
	endDate := time.Now()

	formedQuery := generateQuery(*env, *level, *query)
	resp, err := dd.QueryDatadog(formedQuery, startDate, endDate)
	if err != nil {
		fmt.Printf("ERROR %s\n", err.Error())
		return
	}

	for _,l := range resp.Logs {
	  fmt.Printf("%s : %s\n", l.Content.Timestamp, l.Content.Message)
	}

}
