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

func generateQuery( env string, levels string, query string, all bool) string {

	if all {
		// NO filtering out anything.
		return ""
	}

	//queryTemplate := "@environment:%s status:(%s)"
	queryTemplate := "@EnvSource:%s status:(%s)"

	// status:(info OR warn)
	if strings.TrimSpace(query) != "" {
		queryTemplate = queryTemplate + ` "%s"`
		return fmt.Sprintf(queryTemplate, env,levels,query)
	}

	// check multi level or not.
	levelSplit := strings.Split(levels, ",")
	levelElements := []string{levelSplit[0]}

	if len(levelSplit) > 1 {
		for _,lvl := range levelSplit[1:] {
			levelElements = append(levelElements,"OR")
			levelElements = append(levelElements,lvl)
		}
	}
	return fmt.Sprintf(queryTemplate, env, strings.Join(levelElements, " "))
}

// generateMapForResults map timestamp to list of logs that happened during that minute
// the key (time) is rounded to minute.
func generateMapForResults(resp *models.DatadogQueryResponse) map[time.Time][]models.DataDogLogContent {
	m := make(map[time.Time][]models.DataDogLogContent)

	for _,logEntry := range resp.Logs {

		// rounded to minute
		roundedTime := time.Date( logEntry.Content.Timestamp.Year(), logEntry.Content.Timestamp.Month(),
														  logEntry.Content.Timestamp.Day(), logEntry.Content.Timestamp.Hour(),
														  logEntry.Content.Timestamp.Minute(),0,0,logEntry.Content.Timestamp.Location())

		var logs []models.DataDogLogContent
		var ok bool
		logs, ok = m[roundedTime]
		if !ok {
			logs = []models.DataDogLogContent{}
			m[roundedTime] = logs
		}
		logs = append(logs, logEntry.Content)
		m[roundedTime] = logs
	}

	return m
}

func generateTimeString( t time.Time, loc *time.Location) string {
	lt := t.In(loc)
	dZone,_ := t.Zone()
	lZone,_ := lt.Zone()
	return fmt.Sprintf("%s %s : %s %s", t.Format("2006-01-02 15:04:05"),dZone, lt.Format("2006-01-02 15:04:05"), lZone)
}

// just count for now... needs to add a lot more :)
func displayStats(resp *models.DatadogQueryResponse, startDate time.Time, endDate time.Time) {

	loc,_ := time.LoadLocation("Local")
	logsByTime := generateMapForResults(resp)
	d := time.Date( startDate.Year(), startDate.Month(), startDate.Day(), startDate.Hour(), startDate.Minute(),0,0,
								  startDate.Location())
	for d.Before(endDate) {
		l, ok := logsByTime[d]
		if ok {
			timeString := generateTimeString(d, loc)
			fmt.Printf("%s : %d\n", timeString, len(l))
		}

		d = d.Add(1 * time.Minute)
	}

	counts := len(resp.Logs)
	fmt.Printf("Result count %d\n", counts)
}

func displayResults(resp *models.DatadogQueryResponse, delim bool) {
	loc,_ := time.LoadLocation("Local")
	for _,l := range resp.Logs {
		timeString := generateTimeString(l.Content.Timestamp, loc)
		fmt.Printf("%s : %s\n", timeString, l.Content.Message)
		if delim {
			fmt.Printf("-----------------------------------------------------------------\n")
		}
	}
}

func tailDatadogLogs(dd *pkg.Datadog, startDate time.Time, formedQuery string, delim bool) {

	// startAt is taken from the last search result and passed to the next query
	// It will be blank until we actually GET a result.
	startAt := ""
	var resp *models.DatadogQueryResponse
	var err error
	endDate := time.Now().UTC()

	// tail from this point onwards.
	for {
		if startAt == "" {
			resp, err = dd.QueryDatadog(formedQuery, startDate, endDate)
		} else {
			resp, err = dd.QueryDatadogWithStartAt(formedQuery, startDate, endDate, startAt)
		}

		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			return
		}

		// if results, then display.
		if len(resp.Logs) > 0 {
			displayResults(resp, delim)
			startAt = resp.Logs[ len(resp.Logs)-1].ID
		} else {
			startAt = "" // clear startAt since we have no continuation token :)
		}
	  time.Sleep(30*time.Second)
		startDate = endDate
		endDate = time.Now().UTC()
	}
}

func main() {
	fmt.Printf("So it begins...\n")

	env := flag.String("env", "", "environment: test,stage,prod")
	levels := flag.String("levels", "", "level of logs to query against. info, warn, error. Can be singular or comma separated")
	query := flag.String("query", "", "Part of the query that is NOT specifying level or env.")
	lastNMins := flag.Int("mins", 15, "Last N minutes to be searched")
	stats := flag.Bool("stats", false, "Give summary/stats of logs as opposed to raw logs.")
	delim := flag.Bool("delim", false, "Delimit log entries. Put clear indication between log entries (helpful for spammy logs")
	tail := flag.Bool("tail", false, "Tail the Datadog logs. Will refresh every 30 seconds")
	all := flag.Bool("all", false, "Show all logs, no query to filter out results. Takes priority over all other query related options")

	flag.Parse()

	config := readConfig()
	dd := pkg.NewDatadog(config.DatadogAPIKey, config.DatadogAppKey)
	startDate := time.Now().UTC().Add( time.Duration(-1 * (*lastNMins)) * time.Minute)
	formedQuery := generateQuery(*env, *levels, *query, *all)

	if *tail {
		// just tail constantly. Never exits.
		tailDatadogLogs(dd, startDate, formedQuery, *delim)
		return
	}

	endDate := time.Now().UTC()
	resp, err := dd.QueryDatadog(formedQuery, startDate, endDate)
	if err != nil {
		fmt.Printf("ERROR %s\n", err.Error())
		return
	}
	if *stats {
		// just the stats. :)
		displayStats(resp, startDate, endDate)
	} else {
		displayResults(resp, *delim)
	}
}
