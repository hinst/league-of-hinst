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
	Win bool `json:"win"`

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

func (this *TSessionDetail) FindParticipantByAccountId(accountId int) (result *TSessionDetailParticipant) {
	var pi = this.FindParticipantIdentity(accountId)
	if pi != nil {
		result = this.FindParticipantById(pi.ParticipantId)
	}
	return
}

func (this *TSessionDetail) CheckWinByAccount(accountId int) (result bool) {
	var p = this.FindParticipantByAccountId(accountId)
	if p != nil {
		result = p.Stats.Win
	}
	return
}

func (this *TSessionDetail) FindTeamParticipants(accountId int, myTeam bool) (result []TSessionDetailParticipant) {
	var p = this.FindParticipantByAccountId(accountId)
	if p != nil {
		for _, item := range this.Participants {
			if myTeam {
				if item.TeamId == p.TeamId {
					result = append(result, item)
				}
			} else {
				if item.TeamId != p.TeamId {
					result = append(result, item)
				}
			}
		}
	}
	return
}

func (this *TSessionDetailStats) GetItems() []int {
	return []int{this.Item0, this.Item1, this.Item2, this.Item3, this.Item4, this.Item5, this.Item6}
}

func (this *TSessionDetail) GetItems() (result []int) {
	for _, participant := range this.Participants {
		result = append(result, participant.Stats.GetItems()...)
	}
	return
}

func (this *TSessionDetail) GetTeamItems(accountId int, myTeam bool) (result []int) {
	var participants = this.FindTeamParticipants(accountId, myTeam)
	for _, participant := range participants {
		result = append(result, participant.Stats.GetItems()...)
	}
	return
}
