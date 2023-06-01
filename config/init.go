package config

func init() {
	config = GetConfig("config")
	GetSetting()
}
