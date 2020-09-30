package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kpfaulkner/ddlog/pkg"
	"github.com/kpfaulkner/ddlog/pkg/models"
	"github.com/kpfaulkner/gologmine/pkg/logmine"
	"log"
	"os"
	"os/user"
	"strings"
	"time"
)

// read config from multiple locations.
// first try local dir...
// if fails, try ~/.ddlog/config.json
func readConfig() models.Config {
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

func generateQuery(env string, levels string, query string, all bool) string {

	if all {
		// NO filtering out anything.
		return ""
	}

	// check multi level or not.
	levelSplit := strings.Split(levels, ",")
	levelElements := []string{levelSplit[0]}

	if len(levelSplit) > 1 {
		for _, lvl := range levelSplit[1:] {
			levelElements = append(levelElements, "OR")
			levelElements = append(levelElements, lvl)
		}
	}

	queryTemplate := "@environment:%s status:(%s)"

	if strings.TrimSpace(query) != "" {
		queryTemplate = queryTemplate + ` "%s"`
		return fmt.Sprintf(queryTemplate, env, strings.Join(levelElements, " "), query)
	}

	return fmt.Sprintf(queryTemplate, env, strings.Join(levelElements, " "))
}

func generateTimeString(t time.Time, loc *time.Location, localTimeZone bool) string {
	dZone, _ := t.Zone()
	if localTimeZone {
		lt := t.In(loc)
		lZone, _ := lt.Zone()
		return fmt.Sprintf("%s %s : %s %s", t.Format("2006-01-02 15:04:05.999999"), dZone, lt.Format("2006-01-02 15:04:05.999999"), lZone)
	} else {
		return fmt.Sprintf("%s %s ", t.Format("2006-01-02 15:04:05.999999"), dZone)
	}

}

// just count for now... needs to add a lot more :)
func displayStats(logs []models.DataDogLog, startDate time.Time, endDate time.Time, localTimeZone bool) {

	loc, _ := time.LoadLocation("Local")
	logsByTime := pkg.GroupLogsByMinute(logs)
	d := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), startDate.Hour(), startDate.Minute(), 0, 0,
		startDate.Location())
	for d.Before(endDate) {
		l, ok := logsByTime[d]
		if ok {
			timeString := generateTimeString(d, loc, localTimeZone)
			fmt.Printf("%s : %d\n", timeString, len(l))
		}

		d = d.Add(1 * time.Minute)
	}

	counts := len(logs)
	fmt.Printf("Result count %d\n", counts)
}

func displayResults(logs []models.DataDogLog, delim bool, localTimeZone bool) {
	loc, _ := time.LoadLocation("Local")
	for _, l := range logs {
		timeString := generateTimeString(l.Content.Timestamp, loc, localTimeZone)
		fmt.Printf("%s : %s\n", timeString, l.Content.Message)
		if delim {
			fmt.Printf("-----------------------------------------------------------------\n")
		}
	}
}

// if startAt not found, then just return all logs :)
func filterLogsByStartAt(logs []models.DataDogLog, startAt string) []models.DataDogLog {
	// return them, sorted at least :)
	if startAt == "" {
		return logs
	}

	/*
		sort.Slice( logs, func(i int, j int) bool {
	  	return logs[i].Content.Timestamp.Before(logs[j].Content.Timestamp)
	  })
	*/

	// loop through and only return logs AFTER the startAt is found.
	newLogs := []models.DataDogLog{}
	startAtFound := false
	for _, l := range logs {
		//fmt.Printf("log ID %s : time %s\n", l.ID, l.Content.Timestamp)
		if startAtFound {
			newLogs = append(newLogs, l)
		}

		if l.ID == startAt {
			startAtFound = true
		}
	}

	// if startAt not found, just return original logs.
	if !startAtFound {
		return logs
	}

	return newLogs
}

func tailDatadogLogs(dd *pkg.Datadog, startDate time.Time, formedQuery string, delim bool, localTimeZone bool) {

	// startAt is taken from the last search result and passed to the next query
	// It will be blank until we actually GET a result.
	startAt := ""
	var resp *models.DatadogQueryResponse
	var err error
	endDate := time.Now().UTC()
	lastEndDateWithResults := startDate

	allowedRetries := 5
	// tail from this point onwards.
	for {
		//fmt.Printf("query between %s and %s with startAt %s!\n", startDate, endDate, startAt)
		resp, err = dd.QueryDatadog(formedQuery, startDate, endDate)

		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			// let it continue....  probably just a momentary error, but will give it 5 chances.
			allowedRetries--
			if allowedRetries == 0 {
				return
			}

			fmt.Printf("retries left %d\n", allowedRetries)
		}

		// if results, then display.
		if err == nil && len(resp.Logs) > 0 {
			// if startAt populated, then prune off log entries that we have already displayed.
			logs := filterLogsByStartAt(resp.Logs, startAt)
			//logs := resp.Logs
			if len(logs) > 0 {
				displayResults(logs, delim, localTimeZone)
				startAt = logs[len(logs)-1].ID
				lastEndDateWithResults = logs[len(logs)-1].Content.Timestamp
			}
		}

		time.Sleep(30 * time.Second)
		startDate = lastEndDateWithResults.Add(-5 * time.Second)
		endDate = time.Now().UTC()
	}
}

func main() {
	fmt.Printf("So it begins...\n")

	env := flag.String("env", "prod", "environment: test,stage,prod")
	levels := flag.String("levels", "error", "level of logs to query against. info, warn, error. Can be singular or comma separated")
	query := flag.String("query", "", "Part of the query that is NOT specifying level or env.")
	lastNMins := flag.Int("mins", 15, "Last N minutes to be searched")
	stats := flag.Bool("stats", false, "Give summary/stats of logs as opposed to raw logs.")
	delim := flag.Bool("delim", false, "Delimit log entries. Put clear indication between log entries (helpful for spammy logs")
	tail := flag.Bool("tail", false, "Tail the Datadog logs. Will refresh every 30 seconds")
	all := flag.Bool("all", false, "Show all logs, no query to filter out results. Takes priority over all other query related options")
	local := flag.Bool("local", false, "Shows all log entries in both UTC and local timezones")
	patternLevel := flag.Int("pattern", -1, "Pattern level detection. 0-3 ")  // -1 purely to indicate NOT used.

	flag.Parse()

	config := readConfig()
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
