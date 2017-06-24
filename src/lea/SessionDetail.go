package lea

type TSessionDetail struct {
	GameId       int    `json:"gameId"`
	GameDuration int    `json:"gameDuration"`
	GameType     string `json:"gameType"`

	Teams []TSessionDetailTeam `json:"teams"`

	Participants []TSessionDetailParticipant `json:"participants"`
}

type TSessionDetailTeam struct {
	TeamId int    `json:"teamId"`
	Win    string `json:"win"`
}

type TSessionDetailParticipant struct {
	ParticipantId int `json:"participantId"`
	TeamId        int `json:"teamId"`
	ChampionId    int `json:"championId"`

	Stats TSessionDetailStats `json:"stats"`
}

type TSessionDetailStats struct {
	Item0 int `json:"item0"`
	Item1 int `json:"item1"`
	Item2 int `json:"item2"`
	Item3 int `json:"item3"`
	Item4 int `json:"item4"`
	Item5 int `json:"item5"`
	Item6 int `json:"item6"`
}
