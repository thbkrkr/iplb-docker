package api

import (
	"sync"

	"github.com/gin-gonic/gin"
	iplbapi "github.com/thbkrkr/iplb-docker/iplb"
	"github.com/thbkrkr/iplb-docker/models"
)

type Api struct {
	IPLB *iplbapi.IPLB
}

func (a *Api) Servers(c *gin.Context) {
	servers, err := a.IPLB.GetServers()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, servers)
}

func (a *Api) Backends(c *gin.Context) {
	backends, err := a.IPLB.GetBackends()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, backends)
}

func (a *Api) Frontends(c *gin.Context) {
	frontends, err := a.IPLB.GetFrontends()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, frontends)
}

func (a *Api) Links(c *gin.Context) {
	backends, err := a.IPLB.GetBackends()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	nbBackends := len(backends)

	var wg sync.WaitGroup
	wg.Add(nbBackends)

	links := make(map[int][]models.Link, nbBackends)
	for index, backend := range backends {
		go func(ix int, bid int) {
			defer wg.Done()

			lks, err := a.IPLB.GetLinksByBackendID(bid)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			links[bid] = lks
		}(index, backend.ID)
	}

	wg.Wait()

	c.JSON(200, links)
}
