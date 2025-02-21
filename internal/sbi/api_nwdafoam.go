package sbi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) getNwdafOamRoutes() []Route {
	return []Route{
		{
			Name:    "Health Check",
			Method:  http.MethodGet,
			Pattern: "/",
			APIFunc: func(c *gin.Context) {
				c.String(http.StatusOK, "UPF NWDAF-OAM woking!")
			},
		},
	}
}
