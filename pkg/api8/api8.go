package api8

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"

	"deifzar/reportingm8/pkg/cleanup8"
	"deifzar/reportingm8/pkg/configparser"
	"deifzar/reportingm8/pkg/controller8"
	"deifzar/reportingm8/pkg/db8"
	"deifzar/reportingm8/pkg/log8"
	"deifzar/reportingm8/pkg/orchestrator8"

	"github.com/gin-gonic/gin"
)

type Api8 struct {
	DB     *sql.DB
	Router *gin.Engine
	Config *viper.Viper
	// Orchestrator8  orchestrator8.Orchestrator8Interface
}

func (a *Api8) Init() error {
	// Create configs, log and tmp directories if they don't exist
	if err := os.MkdirAll("configs", 0750); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to create configs directory")
		return err
	}
	if err := os.MkdirAll("log", 0750); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to create log directory")
		return err
	}
	if err := os.MkdirAll("tmp", 0750); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to create tmp directory")
		return err
	}

	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}
	v, err := configparser.InitConfigParser()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Fatal().Msg("Error initialising the config parser.")
		return err
	}

	location := v.GetString("Database.location")
	port := v.GetInt("Database.port")
	schema := v.GetString("Database.schema")
	database := v.GetString("Database.database")
	username := v.GetString("Database.username")
	password := v.GetString("Database.password")

	var db db8.Db8
	db.InitDatabase8(location, port, schema, database, username, password)
	conn, err2 := db.OpenConnection()
	if err2 != nil {
		log8.BaseLogger.Fatal().Msg("Error connecting into DB.")
		return err2
	}

	orchestrator8, err := orchestrator8.NewOrchestrator8()
	if err != nil {
		log8.BaseLogger.Fatal().Msg("Error connecting to the RabbitMQ server.")
		return err
	}
	err = orchestrator8.InitOrchestrator()
	if err != nil {
		log8.BaseLogger.Error().Msg("Error bringing up the RabbitMQ exchanges.")
		return err
	}
	err = orchestrator8.ActivateQueueByService("reportingm8")
	if err != nil {
		log8.BaseLogger.Error().Msg("Error bringing up the RabbitMQ queues for the `reportingm8` service.")
		return err
	}
	err = orchestrator8.ActivateConsumerByService("reportingm8")
	if err != nil {
		log8.BaseLogger.Error().Msg("Error activating consumer with dedicated connection for the `reportingm8` service.")
		return err
	}

	a.DB = conn
	a.Config = v

	// Init schedulerM8
	schedulerM8 := controller8.NewSchedulerm8(a.DB, a.Config)
	err = schedulerM8.InitScheduler()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Fatal().Msg("initializing the scheduler has failed")
		return err
	}

	return nil
}

func (a *Api8) Routes() {
	r := gin.Default()
	// reporting
	schedulerM8 := controller8.NewSchedulerm8(a.DB, a.Config)
	r.GET("/details", schedulerM8.GetSchedulerDetails)
	r.POST("/update", schedulerM8.UpdateScheduler)

	// health checks for kubernetes probes
	r.GET("/health", schedulerM8.HealthCheck)
	r.GET("/ready", schedulerM8.ReadinessCheck)

	a.Router = r
}

func (a *Api8) Run(addr string) {
	a.Router.Run(addr)
}
