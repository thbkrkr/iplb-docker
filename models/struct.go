package models

type Service struct {
	Frontend string
	Backend  string
	Port     int
}

type IPLBService struct {
	Zone  []string `json:"zone"`
	State string   `json:"state"`
}

// Frontend

type AddFrontend struct {
	//AllowedSource string `json:"allowedSource"`
	DefaultFarmID int `json:"defaultBackendId"`
	//DefaulSSlID string `json:"defaulSslId"`
	HSTS bool `json:"hsts"`
	//HTTPHeader string `json:"httpHeader"`
	Port int `json:"port"`
	//RedirectLocation string `json:"redirectLocation"`
	SSL  bool   `json:"ssl"`
	Zone string `json:"zone"`
}

type Frontend struct {
	ID int `json:"frontendId"`
	//AllowedSource string `json:"allowedSource"`
	DefaultFarmID int `json:"defaultFarmId"`
	//DefaulSSlID string `json:"defaulSslId"`
	HSTS bool `json:"hsts"`
	//HTTPHeader string `json:"httpHeader"`
	Port string `json:"port"` // Why not an int here?
	//RedirectLocation string `json:"redirectLocation"`
	//DedicatedIpfo string
	SSL          bool   `json:"ssl"`
	DefaultSslID int    `json:"defaultSslId"`
	Zone         string `json:"zone"`
	Disabled     bool   `json:"disabled"`
	DisplayName  string `json:"displayName"`
}

// Farm

type AddFarm struct {
	Zone string `json:"zone"`
	Port int    `json:"port"`
	//Stickiness string `json:"stickiness"`
	//Balance string `json:"balance"`
	Type  string `json:"type"`
	Probe string `json:"probe"`
}

type Farm struct {
	ID      int    `json:"farmId"`
	Balance string `json:"balance"`
	Zone    string `json:"zone"`
	Name    string `json:"displayName"`
	Port    int    `json:"port"`
	//Probe      string `json:"probe"`
	Stickiness string `json:"stickiness"` // cookie || sourceIp
}

// Server

type AddServer struct {
	Address string `json:"address"`
	Status  string `json:"status"`
	Port    int    `json:"port"`
}

type Server struct {
	ID                   int    `json:"serverId"`
	Address              string `json:"address"`
	Status               string `json:"status"`
	BackendId            int    `json:"backendId"`
	SSL                  bool   `json:"ssl"`
	Cookie               string `json:"cookie"`
	Port                 int    `json:"port"`
	ProxyProtocolVersion string `json:"proxyProtocolVersion"`
	Chain                string `json:"chain"`
	Weight               int    `json:"weight"`
	Backup               bool   `json:"backup"`
	Probe                bool   `json:"probe"`
	DisplayName          string `json:"displayName"`
}

// Route

type Route struct {
	ID          int         `json:"id"`
	Status      string      `json:"status"`
	Weight      int         `json:"weight"`
	Action      RouteAction `json:"action"`
	RouteId     int         `json:"routeId"`
	Rules       []Rule      `json:"rules"`
	DisplayName string      `json:"displayName"`
	FrontendId  int         `json:"frontendId"`
}

type RouteAction struct {
	Target string `json:"target"`
	Status int    `json:"status"`
	Type   string `json:"type"`
}

// Rule

type Rule struct {
	ID      int    `json:"ruleId"`
	Pattern string `json:"pattern"`
	Match   string `json:"match"`
	Negate  bool   `json:"negate"`
	Field   string `json:"field"`
	//SubField string
}

// SSL

type SSL struct {
	ID          int    `json:"id"`
	Serial      string `json:"serial"`
	Subject     string `json:"subject"`
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
	DisplayName string `json:"displayName"`
}
