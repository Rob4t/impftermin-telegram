package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"time"

	"github.com/go-resty/resty/v2"
	tb "gopkg.in/tucnak/telebot.v2"
)

var config = []notifyConfig{
	// {BotToken: "TOKEN", UserID: 123456789, PLZ: "12345", STIKO: "M", Birthdate: "1999-01-01"},
}

const (
	url          = "https://www.impfportal-niedersachsen.de"
	captchaRegex = "<title>[^<>]*Captcha[^<>]*</title>"
)

type (
	notifyConfig struct {
		BotToken  string
		UserID    int
		PLZ       string
		STIKO     string
		Birthdate string
	}
	response struct {
		Name        string `json:"name"`
		VaccineName string `json:"vaccineName"`
		VaccineType string `json:"vaccineType"`
		OutOfStock  bool   `json:"outOfStock"`
	}
	listResponse struct {
		ResultList []response `json:"resultList"`
	}
)

func handleConfig(cfg notifyConfig) {
	b, err := tb.NewBot(tb.Settings{
		Token: cfg.BotToken,
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	t, err := time.Parse("2006-01-02", cfg.Birthdate)
	if err != nil {
		log.Fatal(err)
		return
	}
	ms := t.UnixNano() / 1000000

	// wait between 1 and 10 seconds
	rand.Seed(time.Now().UnixNano())
	waitSec := rand.Intn(10) + 1
	time.Sleep(time.Duration(waitSec) * time.Second)

	// Create a Resty Client
	client := resty.New()
	client.SetHostURL(url+"/portal/rest/").
		SetHeader("Accept", "application/json")

	resp, err := client.R().
		SetResult(&listResponse{}).
		Get(fmt.Sprintf("appointments/findVaccinationCenterListFree/%s?stiko=%s&birthdate=%d", cfg.PLZ, cfg.STIKO, ms))
	result, ok := resp.Result().(*listResponse)
	if !ok {
		b.Send(&tb.User{ID: cfg.UserID}, fmt.Sprintf("Impftermin Script: Couldnt read answer!"))
		return
	}
	re := regexp.MustCompile(captchaRegex)
	if re.Match(resp.Body()) {
		b.Send(&tb.User{ID: cfg.UserID}, fmt.Sprintf("Impftermin Script: Captcha!"))
		return
	}
	for _, res := range result.ResultList {
		if !res.OutOfStock {
			b.Send(&tb.User{ID: cfg.UserID}, fmt.Sprintf("Impftermine!\n%s\n\nImpfstoff: %s\nImpfzentrum: %s\nSTIKO: %s\nGeburtstag: %s", url, res.VaccineName, res.Name, cfg.STIKO, cfg.Birthdate))
		}
	}
}

func main() {
	for _, c := range config {
		handleConfig(c)
	}
}
