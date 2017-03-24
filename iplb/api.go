package iplb

import (
	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

type Api struct {
	IPLBClient *IPLB
}

func call(c *gin.Context, f func() (interface{}, error)) {
	result, err := f()
	if err != nil {
		logrus.Error(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

func (a *Api) Servers(c *gin.Context) {
	call(c, func() (interface{}, error) {
		return a.IPLBClient.GetServers()
	})
}

func (a *Api) Farms(c *gin.Context) {
	call(c, func() (interface{}, error) {
		return a.IPLBClient.GetFarms()
	})
}

func (a *Api) Frontends(c *gin.Context) {
	call(c, func() (interface{}, error) {
		return a.IPLBClient.GetFrontends()
	})
}

func (a *Api) Routes(c *gin.Context) {
	call(c, func() (interface{}, error) {
		return a.IPLBClient.GetRoutes()
	})
}

func (a *Api) SSLs(c *gin.Context) {
	call(c, func() (interface{}, error) {
		return a.IPLBClient.GetSSLs()
	})
}

/*func (a *Api) Links(c *gin.Context) {
	backends, err := a.IPLBClient.GetBackends()
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

			lks, err := a.IPLBClient.GetLinksByBackendID(bid)
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
*/
