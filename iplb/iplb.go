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

func NewIPLB(endpoint string, ak string, as string, ck string, serviceName string) (*IPLB, error) {
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
	address, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Fatal("Fail to get public IP")
		return nil, err
	}

	iplbClient := IPLB{
		ServiceName: serviceName,
		Client:      client,
		Address:     strings.TrimSpace(string(address)),
	}

	return &iplbClient, nil
}

func (i *IPLB) Sync(services []models.Service) {
	logrus.Infof("Sync %d services", len(services))

	kind := "http"
	weight := 100

	for _, service := range services {

		// Server

		server, err := i.GetServerByAddress(i.Address)
		if err != nil {
			logrus.WithError(err).Error("Fail to get server")
			return
		}

		if server == nil {
			logrus.WithField("address", i.Address).Info("Add new server")
			server, err = i.AddServer(i.Address, "active")
			if err != nil {
				logrus.WithError(err).Error("Fail to add server")
				return
			}
		}

		logrus.WithField("zone", server.Zone).Info("Set zone")
		i.Zone = server.Zone

		// Backend

		backend, err := i.GetBackendByPortAndZone(service.Port, i.Zone)
		if err != nil {
			logrus.WithError(err).Error("Fail to get backend")
			return
		}

		if backend == nil {
			logrus.WithField("port", service.Port).Info("Add new backend")
			backend, err = i.AddBackend(service.Port, kind, i.Zone, kind)
			if err != nil {
				logrus.WithError(err).Error("Fail to add backend")
				return
			}
		}

		// Frontend

		frontend, err := i.GetFrontendByBackendID(backend.ID)
		if err != nil {
			logrus.WithError(err).Error("Fail to get frontend")
			return
		}

		if frontend == nil {
			logrus.WithField("port", service.Port).Info("Add new frontend")
			_, err = i.AddFrontend(backend.ID, false, service.Port, false, i.Zone)

			if err != nil {
				logrus.WithError(err).Error("Fail to add frontend")
				return
			}
		}

		// Links

		link, err := i.GetLinkByBackendIDServerIDAndPort(backend.ID, server.ID, service.Port)
		if err != nil {
			logrus.WithError(err).Error("Fail to get link")
			return
		}

		if link == nil {
			logrus.WithField("port", service.Port).Info("Add new link")
			_, err = i.AddLink(backend.ID, false, service.Port, true, server.ID, false, weight)
			if err != nil {
				logrus.WithError(err).Error("Fail to add link")
				return
			}
		}

		logrus.Infof("Service %v registered", service)
	}
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

func (i *IPLB) AddBackend(port int, kind string, zone string, probe string) (*models.Backend, error) {
	var backend = &models.Backend{}
	newBackend := &models.AddBackend{Port: port, Type: kind, Zone: zone, Probe: probe}
	logrus.Warn(newBackend)
	err := i.Client.Post(fmt.Sprintf("/ipLoadbalancing/%s/backend", i.ServiceName), newBackend, backend)
	if err != nil {
		return nil, err
	}

	return backend, nil
}

func (i *IPLB) GetBackendByPortAndZone(port int, zone string) (*models.Backend, error) {
	backends, err := i.GetBackends()
	if err != nil {
		return nil, err
	}

	for _, backend := range backends {
		if backend.Port == port &&
			backend.Zone == zone {
			return &backend, nil
		}
	}

	return nil, nil
}

func (i *IPLB) GetBackends() ([]models.Backend, error) {
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/backend", i.ServiceName), &IDs)
	if err != nil {
		return nil, err
	}

	nbBackends := len(IDs)

	var wg sync.WaitGroup
	wg.Add(nbBackends)

	backends := make([]models.Backend, nbBackends)
	for index, ID := range IDs {
		go func(ix int, id int) {
			defer wg.Done()
			backend, err := i.GetBackendByID(id)
			if err != nil {
				return
			}
			backends[ix] = *backend
		}(index, ID)
	}

	wg.Wait()

	return backends, nil
}

func (i *IPLB) GetBackendByID(ID int) (*models.Backend, error) {
	var backend models.Backend
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/backend/%d", i.ServiceName, ID), &backend)
	if err != nil {
		return nil, err
	}
	return &backend, nil
}

// --

func (i *IPLB) AddFrontend(backendID int, HSTS bool, port int, SSL bool, zone string) (*models.Frontend, error) {
	var frontend = &models.Frontend{}
	newFrontend := &models.AddFrontend{DefaultBackendID: backendID, HSTS: HSTS, Port: port, SSL: SSL, Zone: zone}
	err := i.Client.Post(fmt.Sprintf("/ipLoadbalancing/%s/frontend", i.ServiceName), newFrontend, frontend)
	if err != nil {
		return nil, err
	}

	return frontend, nil
}

func (i *IPLB) GetFrontendByBackendID(backendID int) (*models.Frontend, error) {
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/frontend?defaultBackendId=%d", i.ServiceName, backendID), &IDs)
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
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/frontend", i.ServiceName), &IDs)
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
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/frontend/%d", i.ServiceName, ID), &frontend)
	if err != nil {
		return nil, err
	}
	return &frontend, nil
}

// --

func (i *IPLB) AddServer(address string, status string) (*models.Server, error) {
	var server = &models.Server{}
	newServer := &models.AddServer{Address: address, Status: status}
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
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/server", i.ServiceName), &IDs)
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
			server, err := i.GetServerByID(id)
			if err != nil {
				return
			}
			servers[ix] = *server
		}(index, ID)
	}

	wg.Wait()

	return servers, nil
}

func (i *IPLB) GetServerByID(ID int) (*models.Server, error) {
	var server models.Server
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/server/%d", i.ServiceName, ID), &server)
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// -- Links

func (i *IPLB) AddLink(backendID int, backup bool, port int, probe bool, serverID int, SSL bool, weight int) (*models.Link, error) {
	var link = &models.Link{}
	newLink := &models.AddLink{Backup: backup, Port: port, Probe: probe, ServerID: serverID, SSL: SSL, Weight: weight}
	err := i.Client.Post(fmt.Sprintf("/ipLoadbalancing/%s/backend/%d/server", i.ServiceName, backendID), newLink, link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (i *IPLB) GetLinkByBackendIDServerIDAndPort(backendID int, serverID int, port int) (*models.Link, error) {
	links, err := i.GetLinksByBackendID(backendID)
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if link.Port == port &&
			link.ServerID == serverID {
			return &link, nil
		}
	}

	return nil, nil
}

func (i *IPLB) GetLinksByBackendID(backendID int) ([]models.Link, error) {
	var IDs []int
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/backend/%d/server", i.ServiceName, backendID), &IDs)
	if err != nil {
		return nil, err
	}

	nbLinks := len(IDs)

	var wg sync.WaitGroup
	wg.Add(nbLinks)

	links := make([]models.Link, nbLinks)
	for index, ID := range IDs {
		go func(ix int, id int) {
			defer wg.Done()
			link, err := i.GetLinkByID(backendID, id)
			if err != nil {
				return
			}
			links[ix] = *link
		}(index, ID)
	}

	wg.Wait()

	return links, nil
}

func (i *IPLB) GetLinkByID(backendID int, ID int) (*models.Link, error) {
	var link models.Link
	err := i.Client.Get(fmt.Sprintf("/ipLoadbalancing/%s/backend/%d/server/%d", i.ServiceName, backendID, ID), &link)
	if err != nil {
		return nil, err
	}
	return &link, nil
}
