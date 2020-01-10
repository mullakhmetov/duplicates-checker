package healthcheck

import "github.com/gin-gonic/gin"

import "fmt"

// RegisterHandlers register healtcheck handler in router
func RegisterHandlers(r *gin.Engine, revision string) {
	r.GET("/ping", healthcheck(revision))
}

func healthcheck(revision string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(200, fmt.Sprint("pong ", revision))
	}
}
