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

type AddBackend struct {
	Zone string `json:"zone"`
	Port int    `json:"port"`
	//Stickiness string `json:"stickiness"`
	//Balance string `json:"balance"`
	Type  string `json:"type"`
	Probe string `json:"probe"`
}

type Backend struct {
	ID         int    `json:"id"`
	Zone       string `json:"zone"`
	Name       string `json:"name"`
	Port       int    `json:"port"`
	Stickiness string `json:"stickiness"`
	Balance    string `json:"balance"`
	Type       string `json:"type"`
	Probe      string `json:"probe"`
}

type AddFrontend struct {
	//AllowedSource string `json:"allowedSource"`
	DefaultBackendID int `json:"defaultBackendId"`
	//DefaulSSlID string `json:"defaulSslId"`
	HSTS bool `json:"hsts"`
	//HTTPHeader string `json:"httpHeader"`
	Port int `json:"port"`
	//RedirectLocation string `json:"redirectLocation"`
	SSL  bool   `json:"ssl"`
	Zone string `json:"zone"`
}

type Frontend struct {
	ID int `json:"id"`
	//AllowedSource string `json:"allowedSource"`
	DefaultBackendID int `json:"defaultBackendId"`
	//DefaulSSlID string `json:"defaulSslId"`
	HSTS bool `json:"hsts"`
	//HTTPHeader string `json:"httpHeader"`
	Port string `json:"port"` // Why not an int here?
	//RedirectLocation string `json:"redirectLocation"`
	SSL  bool   `json:"ssl"`
	Zone string `json:"zone"`
}

type AddServer struct {
	Address string `json:"address"`
	Status  string `json:"status"`
}

type Server struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
	Status  string `json:"status"`
	Type    string `json:"type"`
	Zone    string `json:"zone"`
}

type AddLink struct {
	Backup bool `json:"backup"`
	//Chain string `json:"chain"`
	//Cookie string `json:"cookie"`
	Port  int  `json:"port"`
	Probe bool `json:"probe"`
	//ProxyProtocolVersion string `json:"proxyProtocolVersion"`
	ServerID int  `json:"serverId"`
	SSL      bool `json:"ssl"`
	Weight   int  `json:"weight"`
}

type Link struct {
	ID     int  `json:"id"`
	Backup bool `json:"backup"`
	//Chain string `json:"chain"`
	//Cookie string `json:"cookie"`
	Port  int  `json:"port"`
	Probe bool `json:"probe"`
	//ProxyProtocolVersion string `json:"proxyProtocolVersion"`
	ServerID int  `json:"serverId"`
	SSL      bool `json:"ssl"`
	Weight   int  `json:"weight"`
}
