package internal

import (
	"strconv"

	"github.com/d2jvkpn/crons/crons"

	. "github.com/d2jvkpn/go-web/pkg/resp"
	"github.com/gin-gonic/gin"
)

func LoadAPI(rg *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	auth := rg.Group("/api/auth", handlers...)
	task := auth.Group("/manage")

	task.POST("/create", func(ctx *gin.Context) {
		var (
			err  error
			task crons.Task
		)

		if err = ctx.BindJSON(&task); err != nil {
			JSON(ctx, nil, ErrParseFailed(err))
			return
		}

		if err = Manager.AddTask(task); err != nil {
			JSON(ctx, nil, ErrBadRequest(err, "bad request"))
			return
		}

		JSON(ctx, gin.H{"item": task}, nil)
	})

	task.POST("/remove", func(ctx *gin.Context) {
		var (
			id  int
			err error
		)

		id, _ = strconv.Atoi(ctx.DefaultQuery("id", "0"))

		if err = Manager.RemoveTask(id, "api::remove", "todo"); err != nil {
			JSON(ctx, nil, ErrBadRequest(err, "task not found"))
			return
		}

		Ok(ctx)
	})

	task.GET("/find", func(ctx *gin.Context) {
		var (
			id   int
			err  error
			task *crons.Task
		)

		id, _ = strconv.Atoi(ctx.DefaultQuery("id", "0"))

		if task, err = Manager.FindTask(id); err != nil {
			JSON(ctx, nil, ErrBadRequest(err, "task not found"))
		} else {
			JSON(ctx, gin.H{"item": task}, nil)
		}
	})

	task.GET("/find_all", func(ctx *gin.Context) {
		tasks := Manager.CloneTasks(false)
		JSON(ctx, gin.H{"items": tasks}, nil)
	})
}
