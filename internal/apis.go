package internal

import (
	. "github.com/d2jvkpn/go-web/pkg/resp"
	"github.com/gin-gonic/gin"
)

func LoadAPI(rg *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	auth := rg.Group("/api/auth", handlers...)
	task := auth.Group("/task")

	task.GET("/find_all", func(ctx *gin.Context) {
		tasks := _Manager.CloneTasks(false)
		JSON(ctx, gin.H{"items": tasks}, nil)
	})
}
