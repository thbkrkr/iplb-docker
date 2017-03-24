package iplb

import (
	"github.com/Sirupsen/logrus"
	"github.com/thbkrkr/iplb-docker/models"
)

func (i *IPLB) Sync(services []models.Service) {
	logrus.Infof("Sync %d services", len(services))

	kind := "http"
	//weight := 100

	for _, service := range services {

		// Server

		server, err := i.GetServerByAddress(i.Address)
		if err != nil {
			logrus.WithError(err).Error("Fail to get server")
			return
		}

		if server == nil {
			port := 0 // FIXME
			logrus.WithField("zone", i.Zone).WithField("address", i.Address).Info("Add new server")
			server, err = i.AddServer(i.Address, "active", port)
			if err != nil {
				logrus.WithError(err).Error("Fail to add server")
				return
			}
		}

		// farm

		farm, err := i.GetFarmByPortAndZone(service.Port, i.Zone)
		if err != nil {
			logrus.WithError(err).Error("Fail to get farm")
			return
		}

		if farm == nil {
			logrus.WithField("port", service.Port).Info("Add new farm")
			farm, err = i.AddFarm(service.Port, kind, i.Zone, kind)
			if err != nil {
				logrus.WithError(err).Error("Fail to add farm")
				return
			}
		}

		// Frontend

		frontend, err := i.GetFrontendByFarmID(farm.ID)
		if err != nil {
			logrus.WithError(err).Error("Fail to get frontend")
			return
		}

		if frontend == nil {
			logrus.WithField("port", service.Port).Info("Add new frontend")
			_, err = i.AddFrontend(farm.ID, false, service.Port, false, i.Zone)

			if err != nil {
				logrus.WithError(err).Error("Fail to add frontend")
				return
			}
		}

		// Links

		/*link, err := i.GetLinkByfarmIDServerIDAndPort(farm.ID, server.ID, service.Port)
		if err != nil {
			logrus.WithError(err).Error("Fail to get link")
			return
		}

		if link == nil {
			logrus.WithField("port", service.Port).Info("Add new link")
			_, err = i.AddLink(farm.ID, false, service.Port, true, server.ID, false, weight)
			if err != nil {
				logrus.WithError(err).Error("Fail to add link")
				return
			}
		}*/

		logrus.Infof("Service %v registered", service)
	}
}
