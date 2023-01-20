package slack_sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/razorcorp/bfc-bin-collection-notifier/bfc-api"
	"log"
	"net/http"
)

type (
	Webhook string
	Blocks  struct {
		Blocks []Block `json:"blocks"`
	}

	Block struct {
		TypeBlock TypeBlock       `json:"type"`
		Text      *TextBlock      `json:"text,omitempty"`
		Accessory *AccessoryBlock `json:"accessory,omitempty"`
	}

	TypeBlock string

	TextBlock struct {
		ObjectType TypeBlock `json:"type"`
		Text       string    `json:"text"`
	}

	AccessoryBlock struct {
		ObjectType TypeBlock `json:"type"`
		ImageUrl   string    `json:"image_url"`
		AltText    string    `json:"alt_text"`
	}
)

func (h *Webhook) SendMessage(data bfc_api.DataModel) error {
	log.Println("sending schedule to Slack")
	var slack = new(Blocks)
	slack.buildMessage(data)
	return slack.send(*h)
}

func (b *Blocks) buildMessage(data bfc_api.DataModel) {
	log.Println("building message body")
	b.Blocks = append(b.Blocks,
		Block{
			TypeBlock: "header",
			Text: &TextBlock{
				ObjectType: "plain_text",
				Text:       data.Title,
			},
		},
		Block{
			TypeBlock: "divider",
		},
	)

	for _, collection := range data.Collections {
		dateDiff := collection.Date.Diff()
		if dateDiff > 0 && dateDiff <= 7 {
			b.Blocks = append(b.Blocks,
				Block{
					TypeBlock: "section",
					Text: &TextBlock{
						ObjectType: "mrkdwn",
						Text: fmt.Sprintf("*%s* \n%s",
							collection.CollectionType, collection.UpcomingCollections[0]),
					},
					Accessory: &AccessoryBlock{
						ObjectType: "image",
						ImageUrl:   fmt.Sprintf("%s%s", data.BaseUrl, collection.Icon),
						AltText:    "Icon",
					},
				},
				Block{TypeBlock: "divider"},
			)
		}
	}
}

func (b *Blocks) send(url Webhook) error {
	log.Println("sending api call to Slack")
	payload, payloadErr := b.toJson()
	if payloadErr != nil {
		return nil
	}

	//log.Printf("%#v", string(payload))

	req, _ := http.NewRequest("POST", string(url), bytes.NewBuffer(payload))
	req.Header.Add("Content-type", "application/json")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body interface{}
	if resp.StatusCode >= 400 {
		_ = json.NewDecoder(resp.Body).Decode(&body)
		log.Printf("%#v", body)
		return errors.New("failed to send Slack notification")
	}
	log.Println("Slack notification sent")
	return nil
}

func (b *Blocks) toJson() ([]byte, error) {
	log.Println("json marshalling Slack payload")
	output, err := json.Marshal(b)
	if err != nil {
		log.Println("failed to parse Slack payload")
		return nil, err
	}

	return output, nil
}
