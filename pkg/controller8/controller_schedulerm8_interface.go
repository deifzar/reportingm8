package controller8

import "github.com/gin-gonic/gin"

type Scheduler8Interface interface {
	InitScheduler() error
	UpdateScheduler(*gin.Context)
	GetSchedulerDetails(*gin.Context)
}
