package conf

var dev1Config *Config = &Config{
	Listen:           "0.0.0.0:10001",
	GameContractPath: "./desc/satoshicard_release_desc.json",
	LockContractPath: "./desc/satoshicard_timelock_release_desc.json",
	Key:              "ed909bc8d0b35d622a4c3b0c700fce4f1472c533289d5127a782c09c669fb1d7",
	RpcClientConfig: &RpcClientConfig{
		Host:     "127.0.0.1:19002",
		Username: "regtest",
		Password: "123",
	},
}
