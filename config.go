package majsoul

// ServerAddress majsoul server config
type ServerAddress struct {
	ServerAddress  string `json:"serverAddress"`
	GatewayAddress string `json:"gatewayAddress"`
	GameAddress    string `json:"gameAddress"`
}

var ServerAddressList = []*ServerAddress{
	{
		ServerAddress:  "https://game.maj-soul.net",
		GatewayAddress: "wss://gateway-hw.maj-soul.net/gateway",
		GameAddress:    "wss://gateway-hw.maj-soul.com/game-gateway",
	},
	{
		ServerAddress:  "https://game.maj-soul.com",
		GatewayAddress: "wss://gateway-sy.maj-soul.com/gateway",
		GameAddress:    "wss://gateway-sy.maj-soul.com/game-gateway",
	},
}
