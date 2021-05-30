package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"github.com/go-resty/resty/v2"
	"github.com/peterbourgon/diskv"
	tb "gopkg.in/tucnak/telebot.v2"
)

var config = []notifyConfig{
	/*
	{
		BotToken:      "TOKEN",
		ChatIDs:       []int64{-1234},
		ErrorChatIDs:  []int64{-1234},
		PLZ:           "31245",
		STIKO:         "M",
		Birthdate:     "1966-05-03",
		City:          "Hannover",
		FirstName:     "Max",
		LastName:      "Mustermann",
		Gender:        "M",
		Phone:         "012345567",
		StreetName:    "Testweg",
		StreetNumber:  "1",
		IndicationMed: true,
		Email:         "email@example.com",
	},
	*/
}

const (
	url                   = "https://www.impfportal-niedersachsen.de"
	captchaRegex          = "<title>[^<>]*Captcha[^<>]*</title>"
	foundText             = "❗Impftermine❗\n{PortalUrl}\n\nImpfstoff: {VaccineName}\nImpfzentrum: {VaccineCenter}\nSTIKO: {STIKO}\nGeburtstag: {Birthdate}"
	foundAppointment      = "✅Impftermin \n\nSTIKO: {STIKO}\nGeburtstag: {Birthdate}"
	requestInfos          = "Birthdate: {Birthdate}\nPLZ: {PLZ}\nSTIKO: {STIKO}"
	errStatus             = "❌ Impftermin Script: Invalid http status!"
	errReadAnswer         = "❌ Impftermin Script: Couldnt read answer!\n" + requestInfos
	errCaptcha            = "❌ Impftermin Script: Captcha!\n" + requestInfos
	errResultLen          = "❌ Impftermin Script: Result list empty!\n" + requestInfos
	errRegisterCustomer   = "❌ Impftermin Script: Error registering customer!" + requestInfos
	errGetFreeAppointment = "❌ Impftermin Script: Error getting free appointments!" + requestInfos
	errBookingAppointment = "❌ Impftermin Script: Error booking appointments!" + requestInfos
	errRenewToken         = "❌ Impftermin Script: Error renew token!"

	dateTimeLayoutISO       = "2006-01-02 15:04"
	renewTokenTelegramToken = "TOKEN"
	renewTokenErrChatID     = -1234
	renewTokenChatID        = -1234

	passKey         = "Impfscript-Global"
	passServiceName = "Impfscript"
	diskvFolderName = "impfscript-data"
)

type (
	notifyConfig struct {
		BotToken      string
		ChatIDs       []int64
		ErrorChatIDs  []int64
		PLZ           string
		STIKO         string
		Birthdate     string
		City          string
		Email         string
		FirstName     string
		LastName      string
		Gender        string
		Phone         string
		StreetName    string
		StreetNumber  string
		AgeIndication bool
		IndicationMed bool
		IndicationJob bool
	}
	notifyExecutor struct {
		diskv               *diskv.Diskv
		restyClient         *resty.Client
		bot                 *tb.Bot
		cfg                 notifyConfig
		keyring             keyring.Keyring
		firstAppointment    appointmentReserveRequestEntry
		secondAppointment   appointmentReserveRequestEntry
		customerPk          int64
		customerCode        string
		interval1To2        int
		vaccinationCenterPK int64
	}
)

func newExecutor(cfg notifyConfig) (ne *notifyExecutor, err error) {
	ne = &notifyExecutor{
		cfg: cfg,
	}
	var val keyring.Item
	var token string
	ne.keyring, _ = keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{
			keyring.PassBackend,
		},
		ServiceName: passServiceName,
		PassDir:     "/root/.password-store",
	})
	val, err = ne.keyring.Get(passKey)
	token = string(val.Data)
	ne.restyClient = resty.New()
	ne.restyClient.SetHostURL(url+"/portal/rest/").
		SetHeader("Accept", "application/json").
		SetHeader("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0").
		SetHeader("Authorization", token)
	ne.bot, err = tb.NewBot(tb.Settings{
		Token: cfg.BotToken,
	})
	flatTransform := func(s string) []string { return []string{} }
	ne.diskv = diskv.New(diskv.Options{
		BasePath:     diskvFolderName,
		Transform:    flatTransform,
		CacheSizeMax: 1024 * 1024,
	})
	return
}

func replaceText(text string, cfg notifyConfig, res *availableResponse) string {
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

func birthdateToISO8601DateTime(date string) (formated string) {
	tz, _ := time.LoadLocation("Local")
	t, err := time.ParseInLocation("2006-01-02", date, tz)
	if err != nil {
		return
	}
	formated = t.UTC().Format("2006-01-02T15:04:05.000Z")
	return
}

func (ne *notifyExecutor) errorOut(msg string) {
	fmt.Println(msg)
	for _, cid := range ne.cfg.ErrorChatIDs {
		ne.bot.Send(&tb.Chat{ID: cid}, msg)
	}
}

func (ne *notifyExecutor) BookAppointment() (err error) {
	req := bookRequest{
		bookRequestEntry{
			AgeIndication: ne.cfg.AgeIndication,
			Appointments: []appointmentReserveRequestEntry{
				ne.firstAppointment,
				ne.secondAppointment,
			},
			AutomaticScheduling: 0,
			Birthdate:           birthdateToISO8601DateTime(ne.cfg.Birthdate),
			CountryCode:         "DE",
			City:                ne.cfg.City,
			CustomerStatus:      1,
			CustomerPK:          ne.customerPk,
			CustomerCode:        ne.customerCode,
			Email:               ne.cfg.Email,
			Email2:              ne.cfg.Email,
			FirstCustomer:       true,
			FirstName:           ne.cfg.FirstName,
			Gender:              ne.cfg.Gender,
			Interval1To2:        ne.interval1To2,
			Job:                 "",
			LastName:            ne.cfg.LastName,
			MedicalIndication:   ne.cfg.IndicationMed,
			JobIndication:       ne.cfg.IndicationJob,
			Mobilephone:         "",
			Phone:               ne.cfg.Phone,
			SendEmail:           true,
			StreetName:          ne.cfg.StreetName,
			StreetNumber:        ne.cfg.StreetNumber,
			VaccinationPermit:   true,
			Wishlist:            "{\"monday1\":true,\"tuesday1\":true,\"wednesday1\":true,\"thursday1\":true,\"friday1\":true,\"saturday1\":true,\"sunday1\":true,\"monday2\":true,\"tuesday2\":true,\"wednesday2\":true,\"thursday2\":true,\"friday2\":true,\"saturday2\":true,\"sunday2\":true}",
			Zipcode:             ne.cfg.PLZ,
		},
	}
	resp, err := ne.restyClient.R().SetResult(&onlySucceededResponse{}).
		SetBody(req).
		Post("appointments/")
	if !resp.IsSuccess() {
		err = errors.New("http status: " + resp.Status())
		return
	}
	re := regexp.MustCompile(captchaRegex)
	if re.Match(resp.Body()) {
		err = errors.New("Captcha!")
		return
	}

	result, ok := resp.Result().(*onlySucceededResponse)
	if !ok {
		err = errors.New("unmarshal error")
		return
	}
	if result.Succeeded {
		// success message
		for _, cid := range ne.cfg.ChatIDs {
			ne.bot.Send(&tb.Chat{ID: cid}, replaceText(foundAppointment+fmt.Sprintf("\n\nErster: %s\nZweiter: %s\nCode: %s", ne.firstAppointment.AppointmentDate, ne.secondAppointment.AppointmentDate, ne.customerCode), ne.cfg, nil))
		}
		ne.diskv.Write(ne.cfg.FirstName+ne.cfg.LastName, []byte{'1'})
	}
	return
}

func (ne *notifyExecutor) ReserveAppointment(appointment time.Time, appointmentPK int64, customerPK int64, vaccinationCenterPK int64, reason string) (resAppointment appointmentReserveRequestEntry, err error) {
	req := appointmentReserveRequest{
		appointmentReserveRequestEntry{
			AppointmentDate:     appointment.UTC().Format("2006-01-02T15:04:05.000Z"),
			AppointmentPK:       appointmentPK,
			CustomerPK:          customerPK,
			Reason:              reason,
			VaccinationCenterPK: vaccinationCenterPK,
		},
	}
	resp, err := ne.restyClient.R().SetResult(&reservedAppointmentsResponse{}).
		SetBody(req).
		Post("appointments/reserve/")
	if !resp.IsSuccess() {
		err = errors.New("http status: " + resp.Status())
		return
	}

	result, ok := resp.Result().(*reservedAppointmentsResponse)
	if !ok {
		err = errors.New("unmarshal failed")
		return
	}
	resAppointment = result.ResultList[0]
	return
}

func (ne *notifyExecutor) GetFreeAppointments() (err error) {
	year, month, _ := time.Now().Date()
	resp, err := ne.restyClient.R().SetResult(&availableAppointmentsResponse{}).
		Get(fmt.Sprintf("appointments/searchBookedAppointments/%d/%s", ne.vaccinationCenterPK, fmt.Sprintf("%d%02d", year, month)))
	if err != nil {
		return
	}
	if !resp.IsSuccess() {
		err = errors.New("http status: " + resp.Status())
		return
	}
	re := regexp.MustCompile(captchaRegex)
	if re.Match(resp.Body()) {
		err = errors.New("Captcha!")
		return
	}

	result, ok := resp.Result().(*availableAppointmentsResponse)
	if !ok {
		err = errors.New("unmarshal failed")
		return
	}

	var firstAppointmentTime time.Time
	aps := result.ResultList[0]
	keys := make([]string, 0, len(result.ResultList[0]))
	for k := range result.ResultList[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, date := range keys {
		count := aps[date]
		lastOne := false
		if count <= 0 {
			continue
		}
		tz, _ := time.LoadLocation("Local")
		t, _ := time.ParseInLocation(dateTimeLayoutISO, date, tz)
		if t.Before(time.Now().Add(24 * time.Hour)) {
			continue
		}
		if ne.firstAppointment.AppointmentPK == 0 {
			fmt.Println(fmt.Sprintf("reserving because count: %d", count))
			ne.firstAppointment, err = ne.ReserveAppointment(t, ne.firstAppointment.AppointmentPK, ne.customerPk, ne.vaccinationCenterPK, "1")
			if err != nil {
				fmt.Println(err)
				break
			}
			if ne.firstAppointment.AppointmentPK == -1 {
				return errors.New("failed to book first appointment")
			}
			fmt.Println(fmt.Sprintf("Reserved appointment on: %s", t))
			firstAppointmentTime = t
		} else {
			if t.Before(firstAppointmentTime.Add(time.Duration(ne.interval1To2) * 24 * time.Hour)) {
				continue
			}
			ne.secondAppointment, err = ne.ReserveAppointment(t, ne.firstAppointment.AppointmentPK, ne.customerPk, ne.vaccinationCenterPK, "2")
			if err != nil {
				fmt.Println(err)
				break
			}
			if ne.secondAppointment.AppointmentPK == -1 {
				return errors.New("failed to book second appointment")
			}
			fmt.Println(fmt.Sprintf("Reserved appointment on: %s", t))
			lastOne = true
		}
		if lastOne {
			break
		}
	}
	return
}

func (ne *notifyExecutor) RegisterCustomer() (err error) {
	req := appointmentRequest{
		appointmentRequestEntry{
			AgeIndication:       ne.cfg.AgeIndication,
			Appointments:        []string{},
			AutomaticScheduling: 0,
			Birthdate:           birthdateToISO8601DateTime(ne.cfg.Birthdate),
			CountryCode:         "DE",
			City:                ne.cfg.City,
			CustomerStatus:      1,
			Email:               ne.cfg.Email,
			Email2:              ne.cfg.Email,
			FirstCustomer:       true,
			FirstName:           ne.cfg.FirstName,
			Gender:              ne.cfg.Gender,
			Job:                 "",
			LastName:            ne.cfg.LastName,
			MedicalIndication:   ne.cfg.IndicationMed,
			JobIndication:       ne.cfg.IndicationJob,
			Mobilephone:         "",
			Phone:               ne.cfg.Phone,
			SendEmail:           true,
			StreetName:          ne.cfg.StreetName,
			StreetNumber:        ne.cfg.StreetNumber,
			VaccinationPermit:   true,
			Zipcode:             ne.cfg.PLZ,
		},
	}
	resp, err := ne.restyClient.R().SetResult(&appointmentListResponse{}).
		SetBody(req).
		Post("appointments/")
	if err != nil {
		return
	}
	if !resp.IsSuccess() {
		err = errors.New("no success fetching appointment list")
		return
	}
	re := regexp.MustCompile(captchaRegex)
	if re.Match(resp.Body()) {
		err = errors.New("Captcha!")
		return
	}

	result, ok := resp.Result().(*appointmentListResponse)
	if !ok {
		err = errors.New("no success appointment list response")
		return
	}
	if len(result.ResultList) == 0 {
		err = errors.New("empty appointment list response")
		return
	}
	ne.customerPk = result.ResultList[0].CustomerPK
	ne.customerCode = result.ResultList[0].CustomerCode
	return
}

func RenewToken(success chan bool) (res bool) {
	defer func() { success <- res }()

	telegramBot, err := tb.NewBot(tb.Settings{
		Token: renewTokenTelegramToken,
	})

	kr, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{
			keyring.PassBackend,
		},
		ServiceName: passServiceName,
		PassDir:     "/root/.password-store",
	})
	if err != nil {
		fmt.Println(err)
		telegramBot.Send(&tb.Chat{ID: renewTokenErrChatID}, replaceText(errRenewToken+"\n\n"+fmt.Sprint(err), notifyConfig{}, nil))
		return
	}
	var ki keyring.Item
	ki, err = kr.Get(passKey)
	if err != nil {
		fmt.Println(err)
		telegramBot.Send(&tb.Chat{ID: renewTokenErrChatID}, replaceText(errRenewToken+"\n\n"+fmt.Sprint(err), notifyConfig{}, nil))
		return
	}
	val := string(ki.Data)

	// wait between 1 and 45 seconds
	rand.Seed(time.Now().UnixNano())
	waitSec := rand.Intn(45) + 1
	time.Sleep(time.Duration(waitSec) * time.Second)

	client := resty.New()
	client.SetHostURL(url+"/portal/rest/").
		SetHeader("Accept", "application/json").
		SetHeader("User-Agent", "User-Agent: Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0").
		SetHeader("Authorization", val)
	resp, err := client.R().SetResult(&renewTokenResponse{}).
		Get("login/renewtoken")
	if err != nil {
		fmt.Println(err)
		telegramBot.Send(&tb.Chat{ID: renewTokenChatID}, replaceText(errRenewToken+"\n\n"+fmt.Sprint(err), notifyConfig{}, nil))
		return
	}
	if !resp.IsSuccess() {
		fmt.Println("No renew token success")
		telegramBot.Send(&tb.Chat{ID: renewTokenChatID}, replaceText(errRenewToken+"\n\nhttp status: "+resp.Status(), notifyConfig{}, nil))
		return
	}
	re := regexp.MustCompile(captchaRegex)
	if re.Match(resp.Body()) {
		telegramBot.Send(&tb.Chat{ID: renewTokenErrChatID}, replaceText(errRenewToken+"\n\nCaptcha!", notifyConfig{}, nil))
		return
	}

	result, ok := resp.Result().(*renewTokenResponse)
	if !ok {
		fmt.Println("No renew token result")
		telegramBot.Send(&tb.Chat{ID: renewTokenErrChatID}, replaceText(errRenewToken+"\n\nfail to unmarshal renewtoken response", notifyConfig{}, nil))
		return
	}
	if result.Status != "OK" {
		fmt.Println("No renew token success")
		telegramBot.Send(&tb.Chat{ID: renewTokenChatID}, replaceText(errRenewToken+"\n\nstatus was not ok", notifyConfig{}, nil))
		return
	}
	err = kr.Set(keyring.Item{
		Key:  passKey,
		Data: []byte(result.JWTToken),
	})
	if err != nil {
		fmt.Println(err)
		telegramBot.Send(&tb.Chat{ID: renewTokenChatID}, replaceText(errRenewToken+"\n\n"+fmt.Sprint(err), notifyConfig{}, nil))
	}
	res = true
	return
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

	resp, err := ne.restyClient.R().
		SetResult(&availableListResponse{}).
		Get(fmt.Sprintf("appointments/findVaccinationCenterListFree/%s?stiko=%s&birthdate=%d", ne.cfg.PLZ, ne.cfg.STIKO, ms))
	if !resp.IsSuccess() {
		ne.errorOut(replaceText(errStatus+fmt.Sprintf("\nStatus: %d", resp.StatusCode())+"\n"+requestInfos, ne.cfg, nil))
		return
	}

	result, ok := resp.Result().(*availableListResponse)
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
			ne.vaccinationCenterPK = res.VaccinationCenterPk
			ne.interval1To2 = res.Interval1To2
			_, err = ne.diskv.Read(ne.cfg.FirstName + ne.cfg.LastName)
			if err != nil {
				if err := ne.RegisterCustomer(); err != nil {
					ne.errorOut(replaceText(errRegisterCustomer+"\n"+fmt.Sprint(err), ne.cfg, nil))
				} else {
					if err := ne.GetFreeAppointments(); err != nil {
						ne.errorOut(replaceText(errGetFreeAppointment+"\n"+fmt.Sprint(err), ne.cfg, nil))
					} else {
						if err := ne.BookAppointment(); err != nil {
							ne.errorOut(replaceText(errBookingAppointment+"\n"+fmt.Sprint(err), ne.cfg, nil))
						}
					}
				}
			}
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
	chans = append(chans, make(chan bool))
	go RenewToken(chans[0])
	exit := 0
	for i, c := range config {
		exec, err := newExecutor(c)
		if err != nil {
			fmt.Println(err)
			exit = 1
			continue
		}
		chans = append(chans, make(chan bool))
		go exec.Run(chans[i+1])
	}
	for _, ch := range chans {
		if res := <-ch; !res {
			exit = 1
		}
	}
	os.Exit(exit)
}
