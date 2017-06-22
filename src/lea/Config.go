package lea

type TConfig struct {
	ApiKey string
	RootURL string
	SummonerId string
	AccountId string
}

func (this *TConfig) Create() *TConfig {
	return this
}