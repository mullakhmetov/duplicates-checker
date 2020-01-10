package healthcheck

import "github.com/gin-gonic/gin"

// RegisterHandlers register healtcheck handler in router
func RegisterHandlers(r *gin.Engine) {
	r.GET("/ping", healthcheck)
}

func healthcheck(c *gin.Context) {
	c.String(200, "pong")
}
