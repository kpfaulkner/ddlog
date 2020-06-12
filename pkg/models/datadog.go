package models

import "time"

type DataDogLogContent struct {
	Timestamp  time.Time `json:"timestamp"`
	Tags       []string  `json:"tags"`
	Attributes struct {
		CustomAttribute int `json:"customAttribute"`
		Duration        int `json:"duration"`
	} `json:"attributes"`
	Host    string `json:"host"`
	Service string `json:"service"`
	Message string `json:"message"`
}

type DatadogQueryResponse struct {
	Logs []struct {
		ID      string `json:"id"`
		Content DataDogLogContent `json:"content"`
	} `json:"logs"`
	NextLogID string `json:"nextLogId"`
	Status    string `json:"status"`
}


type DatadogQueryRequest struct {
	Query string `json:"query"`
	Time  struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"time"`
	Sort  string `json:"sort"`
	Limit int    `json:"limit"`
}

func GenerateDatadogQuery( query string, from time.Time, to time.Time ) DatadogQueryRequest {
	q := DatadogQueryRequest{}
	q.Query = query
  q.Time.From = from.Format("2006-01-02T15:04:05Z")
  q.Time.To = to.Format("2006-01-02T15:04:05Z")
  q.Sort = "asc"
  q.Limit = 1000
	return q
}
