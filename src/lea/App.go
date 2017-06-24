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
	WriteLog("& Me=Soraka: " + strconv.Itoa(len(details)) + " " + strconv.Itoa(int(sorakaRatio*100)) + "% ; everything below only considers games where me=Soraka")

	var fReport = func(text string, f func(a TSessionDetail) bool) []TSessionDetail {
		var filtered = FilterSessionDetail(details, f)
		var count = len(filtered)
		var ratio = float32(count) / float32(baseCount)
		WriteLog(text + ": " + IntToStr(count) + " " + IntToStr(int(ratio*100)) + "%")
		return filtered
	}

	fReport("Me as Soraka wins", func(a TSessionDetail) bool {
		return a.CheckWinByAccount(this.AccountId)
	})
	WriteLog("")
	fReport("Had either of Executioner, Mortal or Morello", func(a TSessionDetail) bool {
		var items = a.GetItems()
		return CheckIntArrayContains(items, ExecutionerItemId_v7) ||
			CheckIntArrayContains(items, MortalItemId_v7) ||
			CheckIntArrayContains(items, MorelloItemId_v7)
	})
	fReport("Had either of Executioner or Mortal", func(a TSessionDetail) bool {
		var items = a.GetItems()
		return CheckIntArrayContains(items, ExecutionerItemId_v7) ||
			CheckIntArrayContains(items, MortalItemId_v7)
	})

	WriteLog("")
	var countOfTrifectaMyTeam = len(FilterSessionDetail(details, func(a TSessionDetail) bool {
		var items = a.GetTeamItems(this.AccountId, true)
		return CheckIntArrayContains(items, ExecutionerItemId_v7) ||
			CheckIntArrayContains(items, MortalItemId_v7) ||
			CheckIntArrayContains(items, MorelloItemId_v7)
	}))
	var countOfTrifectaEnemyTeam = len(FilterSessionDetail(details, func(a TSessionDetail) bool {
		var items = a.GetTeamItems(this.AccountId, false)
		return CheckIntArrayContains(items, ExecutionerItemId_v7) ||
			CheckIntArrayContains(items, MortalItemId_v7) ||
			CheckIntArrayContains(items, MorelloItemId_v7)
	}))
	WriteLog("Had Ex-er, Mortal or Morello on: my team " + IntToStr(countOfTrifectaMyTeam) + ", enemy team " + IntToStr(countOfTrifectaEnemyTeam))
	var countOfTrifectalessEnemy = baseCount - countOfTrifectaEnemyTeam

	var countOfExerMyTeam = len(FilterSessionDetail(details, func(a TSessionDetail) bool {
		var items = a.GetTeamItems(this.AccountId, true)
		return CheckIntArrayContains(items, ExecutionerItemId_v7) ||
			CheckIntArrayContains(items, MortalItemId_v7)
	}))
	var countOfExerEnemyTeam = len(FilterSessionDetail(details, func(a TSessionDetail) bool {
		var items = a.GetTeamItems(this.AccountId, false)
		return CheckIntArrayContains(items, ExecutionerItemId_v7) ||
			CheckIntArrayContains(items, MortalItemId_v7)
	}))
	WriteLog("Had Ex-er or Mortal on: my team " + IntToStr(countOfExerMyTeam) + ", enemy team " + IntToStr(countOfExerEnemyTeam))
	var countOfExerlessEnemy = baseCount - countOfExerEnemyTeam

	WriteLog("")
	var winsVersusTrifecta = len(FilterSessionDetail(details, func(a TSessionDetail) bool {
		var items = a.GetTeamItems(this.AccountId, false)
		var hadTrifecta = CheckIntArrayContains(items, ExecutionerItemId_v7) || CheckIntArrayContains(items, MortalItemId_v7) || CheckIntArrayContains(items, MorelloItemId_v7)
		return hadTrifecta && a.CheckWinByAccount(this.AccountId)
	}))
	WriteLog("Wins against trifecta on enemy team " + IntToStr(winsVersusTrifecta) + " of " + IntToStr(countOfTrifectaEnemyTeam) + ", ratio=" + RatioToStr(winsVersusTrifecta, countOfTrifectaEnemyTeam) + "%")

	var winsWithTrifecta = len(FilterSessionDetail(details, func(a TSessionDetail) bool {
		var items = a.GetTeamItems(this.AccountId, true)
		var hadTrifecta = CheckIntArrayContains(items, ExecutionerItemId_v7) || CheckIntArrayContains(items, MortalItemId_v7) || CheckIntArrayContains(items, MorelloItemId_v7)
		return hadTrifecta && a.CheckWinByAccount(this.AccountId)
	}))
	WriteLog("Wins with trifecta on my team " + IntToStr(winsWithTrifecta) + " of " + IntToStr(countOfTrifectaMyTeam) + ", ratio=" + RatioToStr(winsWithTrifecta, countOfTrifectaMyTeam) + "%")
	WriteLog("Where trifecta = having either Executioner, Mortal or Morello.")
	var winsVersusTrifectless = len(FilterSessionDetail(details, func(a TSessionDetail) bool {
		var items = a.GetTeamItems(this.AccountId, false)
		var hadTrifecta = CheckIntArrayContains(items, ExecutionerItemId_v7) || CheckIntArrayContains(items, MortalItemId_v7) || CheckIntArrayContains(items, MorelloItemId_v7)
		return false == hadTrifecta && a.CheckWinByAccount(this.AccountId)
	}))
	WriteLog("Wins agains trifectaless enemy: " + IntToStr(winsVersusTrifectless) + " of " + IntToStr(countOfTrifectalessEnemy) + " ratio=" + RatioToStr(winsVersusTrifectless, countOfTrifectalessEnemy) + "%" +
		" delta=" + IntToStr(winsVersusTrifectless-countOfTrifectalessEnemy/2))

	WriteLog("")
	var winsVersusExer = len(FilterSessionDetail(details, func(a TSessionDetail) bool {
		var items = a.GetTeamItems(this.AccountId, false)
		var hadExer = CheckIntArrayContains(items, ExecutionerItemId_v7) || CheckIntArrayContains(items, MortalItemId_v7)
		return hadExer && a.CheckWinByAccount(this.AccountId)
	}))
	WriteLog("Wins against Ex-er or Mortal on enemy team " + IntToStr(winsVersusExer) + " of " + IntToStr(countOfExerEnemyTeam) + ", ratio=" + RatioToStr(winsVersusExer, countOfExerEnemyTeam) + "%")
	var winsWithExer = len(FilterSessionDetail(details, func(a TSessionDetail) bool {
		var items = a.GetTeamItems(this.AccountId, true)
		var hadExer = CheckIntArrayContains(items, ExecutionerItemId_v7) || CheckIntArrayContains(items, MortalItemId_v7)
		return hadExer && a.CheckWinByAccount(this.AccountId)
	}))
	WriteLog("Wins with Ex-er or Mortal on my team " + IntToStr(winsWithExer) + " of " + IntToStr(countOfExerMyTeam) + ", ratio=" + RatioToStr(winsWithExer, countOfExerMyTeam) + "%")
	var winsVersusExerless = len(FilterSessionDetail(details, func(a TSessionDetail) bool {
		var items = a.GetTeamItems(this.AccountId, false)
		var hadTrifecta = CheckIntArrayContains(items, ExecutionerItemId_v7) || CheckIntArrayContains(items, MortalItemId_v7)
		return false == hadTrifecta && a.CheckWinByAccount(this.AccountId)
	}))
	WriteLog("Win agains exerless enemy: " + IntToStr(winsVersusExerless) + " of " + IntToStr(countOfExerlessEnemy) + " ratio=" + RatioToStr(winsVersusExerless, countOfExerlessEnemy) + "%" +
		" delta=" + IntToStr(winsVersusExerless-countOfExerlessEnemy/2))
	WriteLog("Where exerless means: had neither of: Executioner, Mortal")
}
