package routes

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRegisterMediaRoutesOnlyExposeCommonStyleVideoGenerations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &handler.Handlers{Media: &handler.MediaHandler{}}
	RegisterMediaRoutes(r, h, func(c *gin.Context) { c.Next() })

	routesFound := map[string]bool{}
	for _, route := range r.Routes() {
		routesFound[route.Method+" "+route.Path] = true
	}
	require.True(t, routesFound["POST /v1/video/generations"])
	require.True(t, routesFound["GET /v1/video/generations/:task_id"])
	require.False(t, routesFound["POST /v1/videos"])
	require.False(t, routesFound["GET /v1/videos/:task_id"])
}
