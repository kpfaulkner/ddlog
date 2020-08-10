package comms

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DatadogComms struct {
	apiKey string
	appKey string
	url    string
}

func NewDatadogComms(apiKey string, appKey string) DatadogComms {
	c := DatadogComms{}
	c.apiKey = apiKey
	c.appKey = appKey

	// TODO(kpfaulkner) read this in from a config or somewhere.
	c.url = "https://api.datadoghq.com/api/v1/logs-queries/list"
	return c
}

func (c DatadogComms) DoPost(body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(body))
	req.Header.Add("DD-API-KEY", c.apiKey)
	req.Header.Add("DD-APPLICATION-KEY", c.appKey)
	req.Header.Add("Content-type", "application/json")

	// yeah, making a new instance each time. Will worry about this later.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error on post %s\n", err.Error())
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}
