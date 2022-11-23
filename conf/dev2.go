package conf

var dev2Config *Config = &Config{
	Listen:           "0.0.0.0:10002",
	GameContractPath: "./desc/satoshicard_release_desc.json",
	LockContractPath: "./desc/satoshicard_timelock_release_desc.json",
	Key:              "324cd0f6aec47537f4f3f439a9f1c906ac54f04a95d1f5731f3d0cee6888507d",
	RpcClientConfig: &RpcClientConfig{
		Host:     "192.168.10.165:19002",
		Username: "regtest",
		Password: "123",
	},
}
