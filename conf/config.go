package conf

import "flag"

const ()

type RpcClientConfig struct {
	Host     string
	Username string
	Password string
}
type Config struct {
	RpcClientConfig  *RpcClientConfig
	Listen           string
	GameContractPath string
	LockContractPath string
	Key              string
	Mode             int64
	Help             bool
}

var gConfig *Config = &Config{
	RpcClientConfig: &RpcClientConfig{},
}

func Init() {
	config := GetConfig()
	flag.BoolVar(&config.Help, "help", false, "print usage")
	flag.StringVar(&config.Listen, "listen", "0.0.0.0:10001", "host listen port")
	flag.StringVar(&config.RpcClientConfig.Host, "rpchost", "127.0.0.1:19002", "rpc host")
	flag.StringVar(&config.RpcClientConfig.Username, "rpcusername", "regtest", "rpc username")
	flag.StringVar(&config.RpcClientConfig.Password, "rpcpassword", "123", "rpc password")
	flag.StringVar(&config.GameContractPath, "gamecontractpath", "./desc/satoshicard_release_desc.json", "game contract path")
	flag.StringVar(&config.LockContractPath, "lockcontractpath", "./desc/satoshicard_timelock_release_desc.json", "lock contract path")
	flag.StringVar(&config.Key, "key", "ed909bc8d0b35d622a4c3b0c700fce4f1472c533289d5127a782c09c669fb1d7", "private key")
	flag.Int64Var(&config.Mode, "mode", 0, "mode , 0 is regualr,1 and 2 both specify the input,only input space is fine")
	flag.Parse()
	return
}

func GetConfig() *Config {
	return gConfig
}
