package setting

var APPSetting = APP{
	Websocket:        new(Websocket),
	HTTP:             new(HTTP),
	Kafka:            new(Kafka),
	Etcd:             new(Etcd),
	Auth:             new(Auth),
	Redis:            new(Redis),
	ServiceDiscovery: new(ServiceDiscovery),
	RPC:              new(RPC),
}

type APP struct {
	Websocket        *Websocket
	HTTP             *HTTP
	Kafka            *Kafka
	Etcd             *Etcd
	Auth             *Auth
	Redis            *Redis
	ServiceDiscovery *ServiceDiscovery
	RPC              *RPC
}

type Websocket struct {
	Port           int
	Timeout        int
	ReadBufferSize int
}

type HTTP struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
}

type Kafka struct {
	SASLUserName string
	SASLPassword string
	BrokerAddr   []string
}

type Etcd struct {
	Addr     []string
	Username string
	Password string
}

type Redis struct {
	Addr     []string
	Username string
	Password string
}

type Auth struct {
	JwtSecret string
}

type ServiceDiscovery struct {
	GatewayServiceName     string
	GatewayServiceEndpoint string
	PushServiceName        string
	PushServiceEndpoint    string
}

type RPC struct {
	PushRPCPort    int32
	GatewayRPCPort int32
}
