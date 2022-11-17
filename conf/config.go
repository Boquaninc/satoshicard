package conf

type Config struct {
	Listen       string
	ContractPath string
}

var gConfig *Config = nil

func Init(env string) {
	switch env {
	case "dev1":
		gConfig = dev1Config
	case "dev2":
		gConfig = dev2Config
	}
}

func GetConfig() *Config {
	return gConfig
}
