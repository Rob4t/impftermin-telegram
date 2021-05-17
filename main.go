package main

import (
	"fmt"
	"math/rand"
	"os"
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
	requestInfos  = "Birthdate: {Birthdate}\nPLZ: {PLZ}\nSTIKO: {STIKO}"
	errStatus     = "❌ Impftermin Script: Invalid http status!"
	errReadAnswer = "❌ Impftermin Script: Couldnt read answer!\n" + requestInfos
	errCaptcha    = "❌ Impftermin Script: Captcha!\n" + requestInfos
	errResultLen  = "❌ Impftermin Script: Result list empty!\n" + requestInfos
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
	notifyExecutor struct {
		bot *tb.Bot
		cfg notifyConfig
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

func newExecutor(cfg notifyConfig) (ne *notifyExecutor, err error) {
	ne = &notifyExecutor{
		cfg: cfg,
	}
	ne.bot, err = tb.NewBot(tb.Settings{
		Token: cfg.BotToken,
	})
	return
}

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

func (ne *notifyExecutor) errorOut(msg string) {
	fmt.Println(msg)
	for _, cid := range ne.cfg.ErrorChatIDs {
		ne.bot.Send(&tb.Chat{ID: cid}, msg)
	}
}

func (ne *notifyExecutor) Run(success chan bool) (res bool) {
	defer func() { success <- res }()

	t, err := time.Parse("2006-01-02", ne.cfg.Birthdate)
	if err != nil {
		fmt.Println(err)
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
		Get(fmt.Sprintf("appointments/findVaccinationCenterListFree/%s?stiko=%s&birthdate=%d", ne.cfg.PLZ, ne.cfg.STIKO, ms))
	if !resp.IsSuccess() {
		ne.errorOut(replaceText(errStatus+fmt.Sprintf("\nStatus: %d", resp.StatusCode())+"\n"+requestInfos, ne.cfg, nil))
		return
	}

	result, ok := resp.Result().(*listResponse)
	if !ok {
		ne.errorOut(replaceText(errReadAnswer, ne.cfg, nil))
		return
	}
	re := regexp.MustCompile(captchaRegex)
	if re.Match(resp.Body()) {
		ne.errorOut(replaceText(errCaptcha, ne.cfg, nil))
		return
	}
	if len(result.ResultList) == 0 {
		ne.errorOut(replaceText(errResultLen, ne.cfg, nil))
	} else {
		fmt.Println(replaceText("request successful: {Birthdate}", ne.cfg, nil))
	}
	for _, res := range result.ResultList {
		if !res.OutOfStock {
			for _, cid := range ne.cfg.ChatIDs {
				ne.bot.Send(&tb.Chat{ID: cid}, replaceText(foundText, ne.cfg, &res))
			}
		}
	}
	res = true
	return
}

func main() {
	var chans []chan bool
	exit := 0
	for i, c := range config {
		exec, err := newExecutor(c)
		if err != nil {
			fmt.Println(err)
			exit = 1
			continue
		}
		chans = append(chans, make(chan bool))
		go exec.Run(chans[i])
	}
	for _, ch := range chans {
		if res := <-ch; !res {
			exit = 1
		}
	}
	os.Exit(exit)
}
