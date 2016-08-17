package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	dockerapi "github.com/fsouza/go-dockerclient"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"github.com/thbkrkr/go-utilz/http"
	"github.com/thbkrkr/iplb-docker/api"
	iplbapi "github.com/thbkrkr/iplb-docker/iplb"
	"github.com/thbkrkr/iplb-docker/models"
)

type Config struct {
	OvhEndpoint          string `envconfig:"OVH_ENDPOINT" default:"ovh-eu"`
	OvhApplicationKey    string `envconfig:"OVH_AK" required:"true"`
	OvhApplicationSecret string `envconfig:"OVH_AS" required:"true"`
	OvhConsumerKey       string `envconfig:"OVH_CK" required:"true"`
	IpLbServiceName      string `envconfig:"OVH_SERVICENAME" required:"true"`
}

const (
	backendLabel  = "iplb.backend"
	frontendLabel = "iplb.frontend.rule"
	portLabel     = "iplb.port"
	syncInterval  = 30
)

var (
	buildDate = "dev"
	gitCommit = "dev"
	name      = "iplb-docker"
)

var (
	docker   *dockerapi.Client
	config   Config
	services = []models.Service{}
	lock     sync.Mutex
)

func main() {
	// Load config
	err := envconfig.Process("IPLB", &config)
	if err != nil {
		logrus.WithError(err).Fatal("Fail to process config")
	}

	// Create Docker client
	docker, err := dockerapi.NewClientFromEnv()
	assert(err, "Fail to create Docker client")

	// Create IPLB client
	iplb, err := iplbapi.NewIPLB(config.OvhEndpoint,
		config.OvhApplicationKey, config.OvhApplicationSecret, config.OvhConsumerKey,
		config.IpLbServiceName)
	assert(err, "Fail to create OVH IPLB client")

	// Get running containers
	containers, err := docker.ListContainers(dockerapi.ListContainersOptions{})
	assert(err, "Fail to list Docker containers")

	// Filter services to register to iplb in containers list
	for _, container := range containers {
		service := Service(container.Labels)
		if service != nil {
			services = append(services, *service)
		}
	}

	// Sync services in IPLB
	iplb.Sync(services)
	quit := make(chan struct{})
	go func() {
		syncTicker := time.NewTicker(time.Duration(syncInterval) * time.Second)
		for {
			select {

			case <-syncTicker.C:
				lock.Lock()
				iplb.Sync(services)
				lock.Unlock()

			case <-quit:
				syncTicker.Stop()
				return
			}
		}
	}()

	// Listen docker events
	go func() {
		events := make(chan *dockerapi.APIEvents)
		assert(docker.AddEventListener(events), "Fail to listen Docker events")

		for msg := range events {
			switch msg.Status {

			// Add service
			case "start":
				service := Service(msg.Actor.Attributes)
				if service != nil {
					addService(*service)
				}

			// Remove service
			case "die":
				service := Service(msg.Actor.Attributes)
				if service != nil {
					removeService(*service)
				}
			}
		}
	}()

	// HTTP API
	API := api.Api{IPLB: iplb}
	http.API(name, buildDate, gitCommit, func(r *gin.Engine) {
		r.GET("/backend", API.Backends)
		r.GET("/frontend", API.Frontends)
		r.GET("/server", API.Servers)
		r.GET("/link", API.Links)
	})

	close(quit)
	logrus.Fatal("Docker event loop closed")
}

func Service(attributes map[string]string) *models.Service {
	port := attributes[portLabel]
	backend := attributes[backendLabel]
	frontend := attributes[frontendLabel]
	if port != "" && backend != "" && frontend != "" {
		portNum, err := strconv.Atoi(port)
		if err != nil {
			logrus.WithError(err).Errorf("Fail to parse port %s for frontend %s", port, frontend)
			return nil
		}
		return &models.Service{Frontend: frontend, Backend: backend, Port: portNum}
	}
	return nil
}

func addService(service models.Service) {
	lock.Lock()
	defer lock.Unlock()

	services = append(services, service)
}

func removeService(service models.Service) {
	lock.Lock()
	defer lock.Unlock()

	for i := 0; i < len(services); i++ {
		if service.Backend == services[i].Backend &&
			service.Frontend == services[i].Frontend &&
			service.Port == services[i].Port {
			services = append(services[:i], services[i+1:]...)
			break
		}

	}
}

func assert(err error, message string) {
	if err != nil {
		logrus.WithError(err).Fatal(message)
	}
}
