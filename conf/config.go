package conf

type RpcClientConfig struct {
	Host     string
	Username string
	Password string
}
type Config struct {
	RpcClientConfig *RpcClientConfig
	Listen          string
	ContractPath    string
	Key             string
}

var gConfig *Config = nil

func Init(env string) {
	switch env {
	case "dev1":
		gConfig = dev1Config
	case "dev2":
		gConfig = dev2Config
	default:
		panic("not support env")
	}
}

func GetConfig() *Config {
	return gConfig
}
