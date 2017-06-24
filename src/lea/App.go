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
	active    int32
	Config    *TConfig
	AccountId int
}

func (this *TApp) Create() *TApp {
	this.active = 1
	return this
}

func (this *TApp) Run() {
	this.ReadConfig()
	var sessions = this.LoadSessionsToDisk()
	var sessionDetails = this.LoadSessionsFromDisk(sessions.Sessions)
	this.AnalyzeSorakaWounds(sessionDetails)
}

func (this *TApp) ReadConfig() {
	var config = (&TConfig{}).Create()
	var data, readFileResult = ioutil.ReadFile("config.json")
	AssertResult(readFileResult)
	var decodeResult = json.Unmarshal(data, config)
	AssertResult(decodeResult)
	this.Config = config
	WriteLog("RootURL: " + this.Config.RootURL)
	this.AccountId, _ = strconv.Atoi(this.Config.AccountId)
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
	var filePath = this.GetSessionFilePath(gameId)
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

func (this *TApp) LoadSessionsToDisk() *TSessionStructs {
	var sessions = this.RequestSessions()
	for sessionIndex, session := range sessions.Sessions {
		if false == CheckFileExists(this.GetSessionFilePath(session.GameId)) {
			WriteLog("Loading session " + strconv.Itoa(sessionIndex) + "/" + strconv.Itoa(len(sessions.Sessions)) + "...")
			this.LoadSessionFileToDisk(session.GameId)
			time.Sleep(2 * time.Second)
		}
		if false == this.GetActive() {
			break
		}
	}
	return sessions
}

func (this *TApp) GetSessionFilePath(gameId int) string {
	return "data/" + strconv.Itoa(gameId) + ".json"
}

func (this *TApp) LoadSessionsFromDisk(sessions []TSessionStruct) (result []TSessionDetail) {
	result = make([]TSessionDetail, 0, len(sessions))
	for _, session := range sessions {
		var filePath = this.GetSessionFilePath(session.GameId)
		var data, readFileResult = ioutil.ReadFile(filePath)
		AssertResult(readFileResult)
		var sessionDetail TSessionDetail
		json.Unmarshal(data, &sessionDetail)
		result = append(result, sessionDetail)
	}
	return
}

func (this *TApp) AnalyzeSorakaWounds(details []TSessionDetail) {
	WriteLog("Total sessions: " + strconv.Itoa(len(details)))
	details = FilterSessionDetail(details, func(a TSessionDetail) bool {
		return a.GameType == "MATCHED_GAME"
	})
	WriteLog("& Ranked: " + strconv.Itoa(len(details)))
	details = FilterSessionDetail(details, func(a TSessionDetail) bool {
		return a.SeasonId == 8
	})
	var countOfSeasonRanked = len(details)
	WriteLog("& In this season: " + strconv.Itoa(countOfSeasonRanked))
	details = FilterSessionDetail(details, func(a TSessionDetail) bool {
		return a.FindChampionId(this.AccountId) == SorakaChampId_v7
	})
	var countOfSeasonRankedSoraka = len(details)
	var baseCount = countOfSeasonRankedSoraka
	var sorakaRatio = float32(countOfSeasonRankedSoraka) / float32(countOfSeasonRanked)
	WriteLog("& Me=Soraka: " + strconv.Itoa(len(details)) + " " + strconv.Itoa(int(sorakaRatio*100)) + "%")
	var countOfWins = len(
		FilterSessionDetail(details, func(a TSessionDetail) bool {
			return a.CheckWinByAccount(this.AccountId)
		}))
	var winRatio = float32(countOfWins) / float32(baseCount)
	WriteLog("Wins: " + IntToStr(countOfWins) + " " + IntToStr(int(winRatio*100)) + "%")
}
