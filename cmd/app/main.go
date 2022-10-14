package main

import (
	"flag"
	"fmt"
	"github.com/kpfaulkner/ddlog/pkg"
	"github.com/kpfaulkner/ddlog/pkg/models"
	"github.com/kpfaulkner/gologmine/pkg/logmine"
	"log"
	"time"
)

type DDLogApp struct {
	dd *pkg.Datadog
	config models.Config
}

func NewDDLogApp() *DDLogApp {
	app := DDLogApp{}
	app.config = pkg.ReadConfig()
	app.dd = pkg.NewDatadog(app.config.DatadogAPIKey, app.config.DatadogAppKey)
}

func main() {
	fmt.Printf("So it begins...\n")

	config := pkg.ReadConfig()
	dd := pkg.NewDatadog(config.DatadogAPIKey, config.DatadogAppKey)


	startDate := time.Now().UTC().Add(time.Duration(-1*(*lastNMins)) * time.Minute)
	formedQuery := generateQuery(*env, *levels, *query, *all)

	if *tail {
		// just tail constantly. Never exits.
		tailDatadogLogs(dd, startDate, formedQuery, *delim, *local)
		return
	}

	logs := []models.DataDogLog{}
	endDate := time.Now().UTC()
	resp, err := dd.QueryDatadog(formedQuery, startDate, endDate)
	if err != nil {
		fmt.Printf("ERROR %s\n", err.Error())
		return
	}

	logs = append(logs, resp.Logs...)
	// now loop until no nextId
	for resp.NextLogID != "" {
		resp, err = dd.QueryDatadogWithStartAt(formedQuery, startDate, endDate, resp.NextLogID)
		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			return
		}
		logs = append(logs, resp.Logs...)
	}

	if *stats {
		// just the stats. :)
		displayStats(logs, startDate, endDate, *local)
	} else {

		if *patternLevel >= 0 && *patternLevel <= 3 {
			generatePatterns(logs, *patternLevel)

		} else {
			displayResults(logs, *delim, *local)
		}
	}
}


func generatePatterns(logs []models.DataDogLog, maxLevel int) {
  lm := logmine.NewLogMine( []float64{0.01,0.1,0.3,0.9})

  logStringSlice := []string{}
  for _,i := range logs {
  	logStringSlice = append(logStringSlice, i.Content.Message)
  }

  err := lm.ProcessLogsFromSlice(logStringSlice, maxLevel)
	if err != nil {
		log.Fatalf("error while processing. %s\n", err.Error())
	}

	// default to simplified output
	//res, err := lm.GenerateFinalOutput(true)
	lm.DisplayFinalOutput(true)

}
