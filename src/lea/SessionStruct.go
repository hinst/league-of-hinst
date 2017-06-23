package lea

type TSessionStruct struct {
	GameId   int    `json:"gameId"`
	Champion int    `json:"champion"`
	Queue    int    `json:"queue"`
	Season   int    `json:"season"`
	Role     string `json:"role"`
	Lane     string `json:"lane"`
}

type TSessionStructs struct {
	Sessions []TSessionStruct `json:"matches"`
}
