package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	tb "gopkg.in/tucnak/telebot.v2"
)

var config = []notifyConfig{
	// {BotToken: "TOKEN", ChatIDs: []int64{123456}, ErrorChatIDs: []int64{123456}, PLZ: "12345", STIKO: "M", Birthdate: "1999-01-01"},
}

const (
	url           = "https://www.impfportal-niedersachsen.de"
	captchaRegex  = "<title>[^<>]*Captcha[^<>]*</title>"
	foundText     = "❗Impftermine❗\n{PortalUrl}\n\nImpfstoff: {VaccineName}\nImpfzentrum: {VaccineCenter}\nSTIKO: {STIKO}\nGeburtstag: {Birthdate}"
	errReadAnswer = "❌ Impftermin Script: Couldnt read answer!"
	errCaptcha    = "❌ Impftermin Script: Captcha!"
)

type (
	notifyConfig struct {
		BotToken     string
		ChatIDs      []int64
		ErrorChatIDs []int64
		PLZ          string
		STIKO        string
		Birthdate    string
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

func replaceText(text string, cfg notifyConfig, res *response) string {
	replStrings := []string{
		"{PortalUrl}", url,
		"{STIKO}", cfg.STIKO,
		"{Birthdate}", cfg.Birthdate,
		"{PLZ}", cfg.PLZ,
	}
	if res != nil {
		replStrings = append(replStrings, []string{
			"{VaccineName}", res.VaccineName,
			"{VaccineCenter}", res.Name,
			"{OutOfStock}", fmt.Sprint(res.OutOfStock),
		}...)
	}
	r := strings.NewReplacer(replStrings...)
	return r.Replace(text)
}

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

	// wait between 1 and 45 seconds
	rand.Seed(time.Now().UnixNano())
	waitSec := rand.Intn(45) + 1
	time.Sleep(time.Duration(waitSec) * time.Second)

	// Create a Resty Client
	client := resty.New()
	client.SetHostURL(url+"/portal/rest/").
		SetHeader("Accept", "application/json").
		SetHeader("User-Agent", "User-Agent: Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0")

	resp, err := client.R().
		SetResult(&listResponse{}).
		Get(fmt.Sprintf("appointments/findVaccinationCenterListFree/%s?stiko=%s&birthdate=%d", cfg.PLZ, cfg.STIKO, ms))
	result, ok := resp.Result().(*listResponse)
	if !ok {
		fmt.Println(replaceText(errReadAnswer, cfg, nil))
		for _, cid := range cfg.ErrorChatIDs {
			b.Send(&tb.Chat{ID: cid}, replaceText(errReadAnswer, cfg, nil))
		}
		return
	}
	re := regexp.MustCompile(captchaRegex)
	if re.Match(resp.Body()) {
		fmt.Println(replaceText(errCaptcha, cfg, nil))
		for _, cid := range cfg.ErrorChatIDs {
			b.Send(&tb.Chat{ID: cid}, replaceText(errCaptcha+"\nRequest: "+fmt.Sprint(resp.Request.Time)+"\nResponse: "+fmt.Sprint(resp.Time()), cfg, nil))
		}
		return
	} else {
		fmt.Println("got response list len: " + fmt.Sprint(len(result.ResultList)))
	}
	for _, res := range result.ResultList {
		if !res.OutOfStock {
			for _, cid := range cfg.ChatIDs {
				b.Send(&tb.Chat{ID: cid}, replaceText(foundText, cfg, &res))
			}
		}
	}
}

func main() {
	for _, c := range config {
		handleConfig(c)
	}
}
