package main

const (
	ENV_DEV1          = "dev1"
	ENV_DEV2          = "dev2"
	ZOKRATES_CMD_PATH = "/Users/linxing/.local/bin/zokrates"
)

type RpcClientConfig struct {
	Host     string
	Username string
	Password string
}

type Config struct {
	RpcClient    *RpcClientConfig
	ContractPath string
	Listen       string
}

var gConfig *Config

var gConfigDev1 *Config = &Config{
	Listen: "0.0.0.0:10001",
}
var gConfigDev2 *Config = &Config{
	Listen: "0.0.0.0:10002",
}

func InitConfig(env string) {
	switch env {
	case ENV_DEV1:
		gConfig = gConfigDev1
	case ENV_DEV2:
		gConfig = gConfigDev2
	default:
		panic("not support env")
	}
}
