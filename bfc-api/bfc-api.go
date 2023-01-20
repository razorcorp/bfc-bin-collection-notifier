package bfc_api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	DataModel struct {
		Title       string
		Collections []Collection `json:"collections"`
		BaseUrl     string
	}
	Collection struct {
		CollectionType      string   `json:"round"`
		Date                Date     `json:"firstDate"`
		UpcomingCollections []string `json:"upcomingCollections"`
		Icon                string   `json:"icon"`
	}

	Date struct {
		Date string `json:"date"`
	}

	Response struct {
		Result   string    `json:"result"`
		Response DataModel `json:"response"`
	}

	Payload struct {
		CodeAction   string    `json:"code_action"`
		CodeParams   Parameter `json:"code_params"`
		ActionCellId string    `json:"action_cell_id"`
		ActionPageId string    `json:"action_page_id"`
	}

	Parameter struct {
		AddressId string `json:"addressId"`
	}
)

func (p *Payload) GetSchedule(url, path string) (*DataModel, error) {
	log.Println("getting schedule")

	payload := strings.NewReader(
		fmt.Sprintf("code_action=%s&code_params=%s&action_cell_id=%s&action_page_id=%s",
			p.CodeAction, p.CodeParams.UrlEscape(), p.ActionCellId, p.ActionPageId))
	log.Println(payload)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/%s", url, path), payload)
	params := req.URL.Query()
	params.Add("widget_action", "handle_event")
	params.Add("webpage_subpage_id", p.ActionPageId)
	req.URL.RawQuery = params.Encode()

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	client := new(http.Client)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("schedule api status: %d", resp.StatusCode)

	log.Println("decoding response")
	data := Response{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&data); decodeErr != nil {
		return nil, decodeErr
	}
	log.Printf("response status: %s", data.Result)
	return &data.Response, nil
}

func (p *Parameter) UrlEscape() string {
	log.Println("url encoding parameters")
	data, _ := json.Marshal(p)
	encoded := url.QueryEscape(string(data))
	log.Printf("encoded value: %s", encoded)
	return encoded
}

func (d *Date) Format() string {
	log.Println("formatting date to YYYY-MM-DD")
	layout := "2006-01-02 15:04:05.999999999"
	date, err := time.Parse(layout, d.Date)
	if err != nil {
		log.Printf("failed to parse date. %s", err)
		return d.Date
	}

	return date.Format("2006-01-02")
}

func (d *Date) Diff() int {
	log.Println("getting date difference between now and next schedule")
	layout := "2006-01-02 15:04:05.999999999"
	now := time.Now()
	date, dateErr := time.Parse(layout, d.Date)
	if dateErr != nil {
		log.Printf("failed to parse date. %s", dateErr)
		return -1
	}
	return int(date.Sub(now).Hours() / 24)
}
