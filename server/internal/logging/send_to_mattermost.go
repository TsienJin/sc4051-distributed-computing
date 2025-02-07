package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"server/internal/vars"
	"sync"
)

type Payload struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

// Function to strip ANSI escape codes from log messages
func stripAnsiCodes(text string) string {
	ansiEscape := regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)
	return ansiEscape.ReplaceAllString(text, "")
}

func NewPayload(text string) *Payload {
	cleanText := stripAnsiCodes(text) // Remove ANSI codes
	return &Payload{
		Username: "🚀 SC4051 Alert Bot",
		Text:     fmt.Sprintf("%s", cleanText), // Format as code block
	}
}

type MatterMostSender struct {
	client *http.Client
}

var (
	sender             *MatterMostSender
	mattMostSenderOnce sync.Once
)

func getSender() *MatterMostSender {
	mattMostSenderOnce.Do(func() {
		sender = &MatterMostSender{
			client: &http.Client{},
		}
	})

	return sender
}

func SendToMatterMost(text string) {

	url := vars.GetStaticEnv().MatterMostWebhook
	if url == "" {
		return
	}

	p := NewPayload(text)
	data, err := json.Marshal(p)
	if err != nil {
		println(err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	_, err = getSender().client.Do(req)
	if err != nil {
		println(err)
		return
	}

}
