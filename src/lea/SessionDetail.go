package lea

type TSessionDetail struct {
	GameId       int    `json:"gameId"`
	GameDuration int    `json:"gameDuration"`
	SeasonId     int    `json:"seasonId"`
	GameType     string `json:"gameType"`

	Teams []TSessionDetailTeam `json:"teams"`

	Participants []TSessionDetailParticipant `json:"participants"`

	ParticipantIdentities []TSessionDetailParticipantIdentity `json:"participantIdentities"`
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
	Win string `json:"win"`

	Item0 int `json:"item0"`
	Item1 int `json:"item1"`
	Item2 int `json:"item2"`
	Item3 int `json:"item3"`
	Item4 int `json:"item4"`
	Item5 int `json:"item5"`
	Item6 int `json:"item6"`
}

type TSessionDetailParticipantIdentity struct {
	ParticipantId int `json:"participantId"`
	Player        struct {
		AccountId int `json:"accountId"`
	} `json:"player"`
}

func FilterSessionDetail(a []TSessionDetail, f func(a TSessionDetail) bool) (result []TSessionDetail) {
	for _, item := range a {
		if f(item) {
			result = append(result, item)
		}
	}
	return
}

func (this *TSessionDetail) FindChampionId(accountId int) (result int) {
	var pi = this.FindParticipantIdentity(accountId)
	if pi != nil {
		var p = this.FindParticipantById(pi.ParticipantId)
		if p != nil {
			result = p.ChampionId
		}
	}
	return
}

func (this *TSessionDetail) FindParticipantIdentity(accountId int) (result *TSessionDetailParticipantIdentity) {
	for _, item := range this.ParticipantIdentities {
		if item.Player.AccountId == accountId {
			result = &item
			break
		}
	}
	return
}

func (this *TSessionDetail) FindParticipantById(participantId int) (result *TSessionDetailParticipant) {
	for _, item := range this.Participants {
		if item.ParticipantId == participantId {
			result = &item
			break
		}
	}
	return
}
