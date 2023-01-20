package main

import (
	bfcApi "github.com/razorcorp/bfc-bin-collection-notifier/bfc-api"
	slackSdk "github.com/razorcorp/bfc-bin-collection-notifier/slack-sdk"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type (
	Config struct {
		Host      string
		Path      string
		PageId    string
		CellId    string
		AddressId string
		SlackHook slackSdk.Webhook
		Cron      string
	}
)

var VERSION = "latest"

func main() {
	log.Println("Application started")
	log.Printf("Version: %s", VERSION)

	config := Config{
		Host: "https://selfservice.mybfc.bracknell-forest.gov.uk",
		Path: "w/webpage/waste-collection-days",
		Cron: "0 1 * * 1",
		//PageId:    "PAG0000570FEFFB1",
		//CellId:    "PCL0003988FEFFB1",
		//AddressId: "490366",
	}

	if val, ok := os.LookupEnv("BFC_PageId"); !ok {
		log.Fatalf("missing environment variable BFC_PageId")
	} else {
		config.PageId = val
	}

	if val, ok := os.LookupEnv("BFC_CellId"); !ok {
		log.Fatalf("missing environment variable BFC_CellId")
	} else {
		config.CellId = val
	}

	if val, ok := os.LookupEnv("BFC_AddressId"); !ok {
		log.Fatalf("missing environment variable BFC_AddressId")
	} else {
		config.AddressId = val
	}

	if val, ok := os.LookupEnv("Slack_Hook"); !ok {
		log.Fatalf("missing environment variable SLACK_HOOK")
	} else {
		config.SlackHook = slackSdk.Webhook(val)
	}

	if val, ok := os.LookupEnv("Cron"); !ok {
		log.Printf("missing environment variable Cron. Using default %s", config.Cron)
	} else {
		config.Cron = val
	}

	log.Printf("%#v", config)

	//getSchedule(config)

	c := cron.New()
	if eId, err := c.AddFunc(config.Cron, func() {
		getSchedule(config)
	}); err != nil {
		log.Fatalf("%#v", err)
	} else {
		log.Printf("PID: %d, job added", eId)
		//log.Printf("System time: %s", time.Now().String())
		//log.Printf("Next job: %s", c.Entry(eId).Next.String())
	}

	log.Println("starting daemon")
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Println("starting cron jobs")
	c.Start()

	<-shutdown
	log.Println("daemon shutting down")

	log.Println("stopping cron jobs")
	c.Stop()

	log.Println("good bye")
}

func getSchedule(config Config) {
	payload := bfcApi.Payload{
		CodeAction:   "find_rounds",
		CodeParams:   bfcApi.Parameter{AddressId: config.AddressId},
		ActionCellId: config.CellId,
		ActionPageId: config.PageId,
	}
	schedule, scheduleErr := payload.GetSchedule(config.Host, config.Path)
	if scheduleErr != nil {
		log.Fatalf("%#v", scheduleErr)
	}

	schedule.BaseUrl = config.Host
	schedule.Title = "Upcoming Collection Schedule "

	if err := config.SlackHook.SendMessage(*schedule); err != nil {
		log.Fatalf("%#v", err)
	}
}
