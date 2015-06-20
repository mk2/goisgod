package goisgod

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/joeshaw/envdecode"
)

type slackT bool

var slack slackT = slackT(true)

func (s slackT) PostToSlack(format string, args ...interface{}) {

	if !s {
		return
	}

	var env struct {
		WebhookURL string `env:"GOISGOD_SLACK_URL,required"`
	}

	envdecode.Decode(&env)

	var (
		v           url.Values = url.Values{}
		payloadText string
	)

	payload := map[string]string{
		"text": fmt.Sprintf(format, args...),
	}

	if b, err := json.Marshal(payload); err == nil {
		payloadText = string(b)
	}

	v.Set("payload", payloadText)

	http.PostForm(env.WebhookURL, v)
}
