package lea

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type TApp struct {
	Config *TConfig
}

func (this *TApp) Create() *TApp {
	return this
}

func (this *TApp) Run() {
	this.ReadConfig()
	this.RequestSessions()
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

func (this *TApp) RequestSessions() {
	var url = this.Config.RootURL +
		"/api/lol/ru/v2.2/matchlist/by-summoner/" +
		this.Config.SummonerId +
		"?api_key=" + this.Config.ApiKey
	var text = this.Get(url)
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
