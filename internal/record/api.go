package record

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterHandlers register record service handlers in router
func RegisterHandlers(r *gin.Engine, service Service) {
	res := resource{service}

	r.GET("/duples/:u1/:u2", res.IsDuple)
}

type resource struct {
	service Service
}

func (r resource) IsDuple(c *gin.Context) {
	u1, err1 := strconv.Atoi(c.Param("u1"))
	u2, err2 := strconv.Atoi(c.Param("u2"))
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User id param should be integer"})
		return
	}

	res, err := r.service.IsDuple(c, UserID(u1), UserID(u2))
	if err != nil {
		// gin.Logger()

	}
	c.JSON(http.StatusOK, gin.H{"dupes": res})
}
