package lea

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type TApp struct {
	active int32
	Config *TConfig
}

func (this *TApp) Create() *TApp {
	this.active = 1
	return this
}

func (this *TApp) Run() {
	this.ReadConfig()
	var sessions = this.RequestSessions()
	var relevantSessions = this.GetRelevantSessions(sessions.Sessions)
	WriteLog("Got sessions " + strconv.Itoa(len(sessions.Sessions)) + " -> " + strconv.Itoa(len(relevantSessions)))
	for sessionIndex, session := range sessions.Sessions {
		WriteLog("Loading session " + strconv.Itoa(sessionIndex) + "/" + strconv.Itoa(len(sessions.Sessions)) + "...")
		this.LoadSessionFileToDisk(session.GameId)
		time.Sleep(2 * time.Second)
		if false == this.GetActive() {
			break
		}
	}
}

func (this *TApp) ReadConfig() {
	var config = (&TConfig{}).Create()
	var data, readFileResult = ioutil.ReadFile("config.json")
	AssertResult(readFileResult)
	var decodeResult = json.Unmarshal(data, config)
	AssertResult(decodeResult)
	this.Config = config
	WriteLog("RootURL: " + this.Config.RootURL)
}

func (this *TApp) RequestSessions() *TSessionStructs {
	var url = this.Config.RootURL +
		"/lol/match/v3/matchlists/by-account/" + this.Config.AccountId +
		"?api_key=" + this.Config.ApiKey
	var text = this.Get(url)
	var sessionList TSessionStructs
	json.Unmarshal(text, &sessionList)
	return &sessionList
}

func (this *TApp) RequestSessionRaw(gameId int) []byte {
	var url = this.Config.RootURL +
		"/lol/match/v3/matches/" + strconv.Itoa(gameId) +
		"?api_key=" + this.Config.ApiKey
	var data = this.Get(url)
	return data
}

func (this *TApp) GetResponse(url string) *http.Response {
	WriteLog("Get " + url)
	var response, responseResult = http.Get(url)
	AssertResult(responseResult)
	return response
}

func (this *TApp) Get(url string) []byte {
	var resp = this.GetResponse(url)
	var data, readResult = ioutil.ReadAll(resp.Body)
	AssertResult(readResult)
	return data
}

func (this *TApp) GetRelevantSessions(a []TSessionStruct) (result []TSessionStruct) {
	for _, session := range a {
		if session.Champion == SorakaChampId_v7 && session.Season == 8 {
			result = append(result, session)
		}
	}
	return
}

func (this *TApp) LoadSessionFileToDisk(gameId int) {
	var data = this.RequestSessionRaw(gameId)
	var filePath = "data/" + strconv.Itoa(gameId) + ".json"
	var writeFileResult = ioutil.WriteFile(filePath, data, os.ModePerm)
	AssertResult(writeFileResult)
}

func (this *TApp) SetActive(a bool) {
	if a {
		atomic.StoreInt32(&this.active, 1)
	} else {
		atomic.StoreInt32(&this.active, 0)
	}
}

func (this *TApp) GetActive() bool {
	return atomic.LoadInt32(&this.active) > 0
}
