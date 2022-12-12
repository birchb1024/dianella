package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/birchb1024/dianella"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type mySlackStep struct {
	dianella.Step
}

func BEGINslack(desc string) *mySlackStep {
	m := mySlackStep{}
	m.Init(&m, desc)
	return &m
}

func (s *mySlackStep) SendSlack(blockFile string) dianella.Stepper {
	// curl -v -H "Content-type: application/json" --data @block.json  -H "Authorization: Bearer REDACTED" -X POST https://slack.com/api/chat.postMessage
	if s.IsFailed() {
		return s
	}

	f, err := os.Open(blockFile)
	if err != nil {
		s.Fail(err.Error())
	}
	defer func() { _ = f.Close() }()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s", s.Flag["postMessageURL"]), f)
	if err != nil {
		s.Fail(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+fmt.Sprintf("%s", s.Flag["bearerToken"]))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.Fail(err.Error())
	}
	if resp.StatusCode != 200 {
		s.Fail(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.Fail(err.Error())
	}

	jsonResponse := map[string]any{}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		s.Fail(err.Error())
	}
	if jsonResponse["ok"] == false {
		s.Fail(string(body))
	}
	defer func() { _ = resp.Body.Close() }()
	return s
}

var slackPostMessageURL string
var slackBearerToken string
var slackChannelName string

func main() {
	flag.StringVar(&slackPostMessageURL, "postMessageURL", "https://slack.com/api/chat.postMessage", "Slack API Bearer Token")
	flag.StringVar(&slackBearerToken, "bearerToken", "", "Slack API Bearer Token")
	flag.StringVar(&slackChannelName, "channelName", "test-slackbot-bill-birch", "SlackBot Channel Name")
	flag.Parse()

	s := BEGINslack("Start notifications")
	slackReportTemplate, _ := s.Sbash("cat slackBlockTemplate.txt")
	s.AND("Read the epoch date and time").
		Set("date", time.Now().Unix()).
		AND("Generate the slack block JSON message").
		Expand(slackReportTemplate, "block.json").
		AND("Send a message to a slack channel")
	s.SendSlack("block.json").
		END()
	s = s
}
