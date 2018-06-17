package Config

type Map map[string]string

type config interface {
	GetConfigValue(value string)
}

func (config Map) GetConfigValue(value string) (string, bool) {
	if val, ok := config[value]; ok {
		return val, true
	}
	return "", false
}

func InitConfig() Map {
	var m = make(map[string]string)
	m["twitch_bot_token"] = ""

	return m
}
