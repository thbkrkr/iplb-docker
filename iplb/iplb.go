package iplb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/ovh/go-ovh/ovh"
	"github.com/thbkrkr/iplb-docker/models"
)

type IPLB struct {
	ServiceName string
	Zone        string
	Address     string
	Client      *ovh.Client
}

func NewIPLB(endpoint string, ak string, as string, ck string, zone string, serviceName string) (*IPLB, error) {
	client, err := ovh.NewClient(endpoint, ak, as, ck)
	if err != nil {
		logrus.WithError(err).Fatal("Fail to create OVH client")
		return nil, err
	}

	resp, err := http.Get("http://ipaddr.ovh")
	if err != nil {
		logrus.WithError(err).Fatal("Fail to get public IP")
		return nil, err
	}
	defer resp.Body.Close()
	publicIP, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Fatal("Fail to get public IP")
		return nil, err
	}
	address := strings.TrimSpace(string(publicIP))
	logrus.Info("IPLB started with current address ", address)

	iplbClient := IPLB{
		Zone:        zone,
		ServiceName: serviceName,
		Address:     address,
		Client:      client,
	}

	return &iplbClient, nil
}

func (i *IPLB) GetService() (*models.IPLBService, error) {
	var service models.IPLBService
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s", i.ServiceName), &service)
	if err != nil {
		logrus.WithError(err).Fatal("Fail to get IPLB service")
		return nil, err
	}

	return &service, nil
}

// --

func (i *IPLB) AddFarm(port int, kind string, zone string, probe string) (*models.Farm, error) {
	var farm = &models.Farm{}
	newFarm := &models.AddFarm{Port: port, Type: kind, Zone: zone, Probe: probe}
	err := i.Client.Post(fmt.Sprintf("/ipLoadbalancing/%s/farm", i.ServiceName), newFarm, farm)
	if err != nil {
		return nil, err
	}

	return farm, nil
}

func (i *IPLB) GetFarmByPortAndZone(port int, zone string) (*models.Farm, error) {
	farms, err := i.GetFarms()
	if err != nil {
		return nil, err
	}

	for _, farm := range farms {
		if farm.Port == port &&
			farm.Zone == zone {
			return &farm, nil
		}
	}

	return nil, nil
}

func (i *IPLB) GetFarms() ([]models.Farm, error) {
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/http/farm", i.ServiceName), &IDs)
	if err != nil {
		return nil, err
	}

	nbfarms := len(IDs)

	var wg sync.WaitGroup
	wg.Add(nbfarms)

	farms := make([]models.Farm, nbfarms)
	for index, ID := range IDs {
		go func(ix int, id int) {
			defer wg.Done()
			farm, err := i.GetFarmByID(id)
			if err != nil {
				logrus.WithError(err).Error("GetFarmByID")
				return
			}
			farms[ix] = *farm
		}(index, ID)
	}

	wg.Wait()

	return farms, nil
}

func (i *IPLB) GetFarmByID(ID int) (*models.Farm, error) {
	var farm models.Farm
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/http/farm/%d", i.ServiceName, ID), &farm)
	if err != nil {
		return nil, err
	}
	return &farm, nil
}

// --

func (i *IPLB) AddFrontend(farmID int, HSTS bool, port int, SSL bool, zone string) (*models.Frontend, error) {
	var frontend = &models.Frontend{}
	newFrontend := &models.AddFrontend{DefaultFarmID: farmID, HSTS: HSTS, Port: port, SSL: SSL, Zone: zone}
	err := i.Client.Post(fmt.Sprintf("/ipLoadbalancing/%s/frontend", i.ServiceName), newFrontend, frontend)
	if err != nil {
		return nil, err
	}

	return frontend, nil
}

func (i *IPLB) GetFrontendByFarmID(farmID int) (*models.Frontend, error) {
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/frontend?defaultfarmId=%d", i.ServiceName, farmID), &IDs)
	if err != nil {
		return nil, err
	}

	if len(IDs) != 1 {
		return nil, nil
	}

	frontend := &models.Frontend{}
	var ID = IDs[0]
	err = i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/frontend/%d", i.ServiceName, ID), frontend)
	if err != nil {
		return nil, err
	}

	return frontend, nil
}

func (i *IPLB) GetFrontends() ([]models.Frontend, error) {
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/http/frontend", i.ServiceName), &IDs)
	if err != nil {
		return nil, err
	}

	nbFrontends := len(IDs)

	var wg sync.WaitGroup
	wg.Add(nbFrontends)

	frontends := make([]models.Frontend, nbFrontends)
	for index, ID := range IDs {
		go func(ix int, id int) {
			defer wg.Done()
			frontend, err := i.GetFrontendByID(id)
			if err != nil {
				logrus.WithError(err).Error("GetFrontendByID")
				return
			}
			frontends[ix] = *frontend
		}(index, ID)
	}

	wg.Wait()

	return frontends, nil
}

func (i *IPLB) GetFrontendByID(ID int) (*models.Frontend, error) {
	var frontend models.Frontend
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/http/frontend/%d", i.ServiceName, ID), &frontend)
	if err != nil {
		return nil, err
	}
	return &frontend, nil
}

// --

func (i *IPLB) AddServer(address string, status string, port int) (*models.Server, error) {
	var server = &models.Server{}
	newServer := &models.AddServer{Address: address, Status: status, Port: port}
	err := i.Client.Post(fmt.Sprintf("/ipLoadbalancing/%s/server", i.ServiceName), newServer, server)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (i *IPLB) GetServerByAddress(address string) (*models.Server, error) {
	var serverIDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/server?address=%s", i.ServiceName, address), &serverIDs)
	if err != nil {
		return nil, err
	}

	if len(serverIDs) != 1 {
		return nil, nil
	}

	server := &models.Server{}
	var serverID = serverIDs[0]
	err = i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/server/%d", i.ServiceName, serverID), server)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (i *IPLB) GetServers() ([]models.Server, error) {
	farms, err := i.GetFarms()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mx sync.Mutex
	wg.Add(len(farms))

	servers := []models.Server{}
	for _, farm := range farms {
		go func(id int) {
			defer wg.Done()
			serversByFarm, err := i.GetServersByfarmID(id)
			if err != nil {
				logrus.WithError(err).Error("GetServersByfarmID")
				return
			}
			mx.Lock()
			servers = append(servers, serversByFarm...)
			mx.Unlock()
		}(farm.ID)
	}

	wg.Wait()

	return servers, nil
}

/*func (i *IPLB) GetServerByID(ID int) (*models.Server, error) {
	var server models.Server
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/server/%d", i.ServiceName, ID), &server)
	if err != nil {
		return nil, err
	}
	return &server, nil
}*/

/*
func (i *IPLB) AddServer(farmID int, backup bool, port int, probe bool, serverID int, SSL bool, weight int) (*models.Server, error) {
	var server = &models.Server{}
	newServer := &models.AddServer{Address: address, Port: port, Probe: probe, ServerID: serverID, SSL: SSL, Weight: weight}
	err := i.Client.Post(fmt.Sprintf("/ipLoadbalancing/%s/farm/%d/server", i.ServiceName, farmID), newServer, server)
	if err != nil {
		return nil, err
	}

	return server, nil
}
*/

func (i *IPLB) GetServerByfarmIDServerIDAndPort(farmID int, serverID int, port int) (*models.Server, error) {
	servers, err := i.GetServersByfarmID(farmID)
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.Port == port &&
			server.ID == serverID {
			return &server, nil
		}
	}

	return nil, nil
}

func (i *IPLB) GetServersByfarmID(farmID int) ([]models.Server, error) {
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/http/farm/%d/server", i.ServiceName, farmID), &IDs)
	if err != nil {
		return nil, err
	}

	nbServers := len(IDs)

	var wg sync.WaitGroup
	wg.Add(nbServers)

	servers := make([]models.Server, nbServers)
	for index, ID := range IDs {
		go func(ix int, id int) {
			defer wg.Done()
			server, err := i.GetServerByID(farmID, id)
			if err != nil {
				logrus.WithError(err).Error("GetServerByID")
				return
			}
			servers[ix] = *server
		}(index, ID)
	}

	wg.Wait()

	return servers, nil
}

func (i *IPLB) GetServerByID(farmID int, ID int) (*models.Server, error) {
	var Server models.Server
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/http/farm/%d/server/%d", i.ServiceName, farmID, ID), &Server)
	if err != nil {
		return nil, err
	}
	return &Server, nil
}

// -- Rules

func (i *IPLB) GetRules(routeID int) ([]models.Rule, error) {
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/http/route/%d/rule", i.ServiceName, routeID), &IDs)
	if err != nil {
		return nil, err
	}

	nbRules := len(IDs)

	var wg sync.WaitGroup
	wg.Add(nbRules)

	rules := make([]models.Rule, nbRules)
	for index, ID := range IDs {
		go func(ix int, id int) {
			defer wg.Done()
			rule, err := i.GetRuleByID(routeID, id)
			if err != nil {
				logrus.WithError(err).Error("GetRuleByID")
				return
			}
			rules[ix] = *rule
		}(index, ID)
	}

	wg.Wait()

	return rules, nil
}

func (i *IPLB) GetRuleByID(routeID int, ID int) (*models.Rule, error) {
	var rule models.Rule
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/http/route/%d/rule/%d", i.ServiceName, routeID, ID), &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// -- Routes

func (i *IPLB) GetRoutes() ([]models.Route, error) {
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/http/route", i.ServiceName), &IDs)
	if err != nil {
		return nil, err
	}

	nbRoutes := len(IDs)

	var wg sync.WaitGroup
	wg.Add(nbRoutes)

	routes := make([]models.Route, nbRoutes)
	for index, ID := range IDs {
		go func(ix int, id int) {
			defer wg.Done()
			ssl, err := i.GetRouteByID(id)
			if err != nil {
				logrus.WithError(err).Error("GetRouteByID")
				return
			}
			routes[ix] = *ssl
		}(index, ID)
	}

	wg.Wait()

	wg.Add(nbRoutes)
	for index, ID := range IDs {
		go func(ix int, id int) {
			defer wg.Done()
			rules, err := i.GetRules(id)
			if err != nil {
				logrus.WithError(err).Error("GetRouteByID")
				return
			}
			routes[ix].Rules = rules
		}(index, ID)
	}

	wg.Wait()

	return routes, nil
}

func (i *IPLB) GetRouteByID(ID int) (*models.Route, error) {
	var route models.Route
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/http/route/%d", i.ServiceName, ID), &route)
	if err != nil {
		return nil, err
	}
	return &route, nil
}

// -- SSLs

func (i *IPLB) GetSSLs() ([]models.SSL, error) {
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/ssl", i.ServiceName), &IDs)
	if err != nil {
		return nil, err
	}

	nbSSLs := len(IDs)

	var wg sync.WaitGroup
	wg.Add(nbSSLs)

	ssls := make([]models.SSL, nbSSLs)
	for index, ID := range IDs {
		go func(ix int, id int) {
			defer wg.Done()
			ssl, err := i.GetSSLByID(id)
			if err != nil {
				logrus.WithError(err).Error("GetSSLByID")
				return
			}
			ssls[ix] = *ssl
		}(index, ID)
	}

	wg.Wait()

	return ssls, nil
}

func (i *IPLB) GetSSLByID(ID int) (*models.SSL, error) {
	var ssl models.SSL
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/ssl/%d", i.ServiceName, ID), &ssl)
	if err != nil {
		return nil, err
	}
	return &ssl, nil
}
