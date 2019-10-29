package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"os"
)

var webhookURL = os.Getenv("WEBHOOK_URL")
var channel = os.Getenv("CHANNEL")

type AlertLevel int

const (
	Danger AlertLevel = iota
	Warn
	Health
)

var colors = map[AlertLevel]string{
	Danger: "#fc2f2f",
	Warn:   "#ffcc14",
	Health: "#27d871",
}

// {
//   "incident": {
//     "incident_id": "f2e08c333dc64cb09f75eaab355393bz",
//     "resource_id": "i-4a266a2d",
//     "resource_name": "webserver-85",
//     "state": "open",
//     "started_at": 1385085727,
//     "ended_at": null,
//     "policy_name": "Webserver Health",
//     "condition_name": "CPU usage",
//     "url": "https://app.google.stackdriver.com/incidents/f333dc64z",
//     "summary": "CPU for webserver-85 is above the threshold of 1% with a value of 28.5%"
//   },
//   "version": 1.1
// }

type Incident struct {
	IncidentID    string `json:"incident_id"`
	ResourceID    string `json:"resource_id"`
	ResourceName  string `json:"resource_name"`
	State         string `json:"state"`
	StartedAt     int64  `json:"started_at"`
	EndedAt       int64  `json:"ended_at"`
	PolicyName    string `json:"policy_name"`
	ConditionName string `json:"condition_name"`
	URL           string `json:"url"`
	Summary       string `json:"summary"`
}

type Alert struct {
	Incident Incident `json:"incident"`
	Version  float64  `json:"version"`
}

// Slack struct - payload parameter of json to post.
type SlackParam struct {
	Text        string       `json:"text"`
	Username    string       `json:"username"`
	IconEmoji   string       `json:"icon_emoji"`
	IconURL     string       `json:"icon_url"`
	Channel     string       `json:"channel"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Color  string  `json:"color"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func AlertToSlack(w http.ResponseWriter, r *http.Request) {
	alert := Alert{}

	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		log.Printf("[error] decode alert error, %s.", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("[alert log] %v", alert)

	var mention = "<!here>"
	var color = colors[Warn]
	if strings.HasPrefix(alert.Incident.ConditionName, "[DANGER]") {
		mention = "<!channel>"
		color = colors[Danger]
	}
	if alert.Incident.State == "closed" {
		color = colors[Health]
	}

	params := SlackParam{
		Text: fmt.Sprintf("%s %s %s",
			mention,
			alert.Incident.Summary,
			alert.Incident.URL,
		),
		Username: "Alert by Stackdriver",
		Channel:  channel,
		Attachments: []Attachment{
			Attachment{
				Color: color,
				Fields: []Field{
					{
						Title: alert.Incident.Summary,
						Value: buildText(alert),
					},
				},
			},
		},
	}

	b, err := json.Marshal(params)
	if err != nil {
		log.Printf("[error] marshal error, %s.", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := http.PostForm(
		webhookURL,
		url.Values{"payload": {string(b)}},
	)
	if err != nil {
		log.Printf("[error] post form error, %s.", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%s", err.Error())

	} else {
		log.Printf("[slack response] %s", body)

	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(b); err != nil {
		log.Printf("[error] write error, %s.", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

var alertTextFmt = "state: %s\nresources_id: %s\nresources_name: %s"

func buildText(alert Alert) string {
	return fmt.Sprintf(alertTextFmt, alert.Incident.State, alert.Incident.ResourceID, alert.Incident.ResourceName)
}
