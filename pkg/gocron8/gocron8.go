package gocron8

import (
	"fmt"
	"sync"
	"time"

	"deifzar/reportingm8/pkg/log8"
	"deifzar/reportingm8/pkg/notification8"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

var once sync.Once

var BaseReportScheduler gocron.Scheduler
var BaseReportJobOptionsEventListener, BaseReportJobOptionsTags gocron.JobOption

func GetGoCron8() (gocron.Scheduler, gocron.JobOption, gocron.JobOption) {
	once.Do(func() {

		location, err := time.LoadLocation("Local")
		if err != nil {
			log8.BaseLogger.Debug().Stack().Msg(err.Error())
			log8.BaseLogger.Fatal().Msg("Error in Cobra CLI execute. Something wrong with setting the `Local` time zone for the GoCron.")
		}

		// create a scheduler
		BaseReportScheduler, err = gocron.NewScheduler(
			gocron.WithLocation(location),
		)
		if err != nil {
			log8.BaseLogger.Debug().Stack().Msg(err.Error())
			log8.BaseLogger.Fatal().Msg("Error in Cobra CLI execute. Something wrong with creating the GoCron.")
		}

		// event listener
		BaseReportJobOptionsEventListener = gocron.WithEventListeners(
			gocron.AfterJobRuns(
				func(jobID uuid.UUID, jobName string) {
					log8.BaseLogger.Debug().Stack().Msgf("The job '%s' (%s) has found errors: %s", jobName, jobID, err)
					log8.BaseLogger.Info().Msgf("The following the GoCron job has finshed: %s(%s).", jobName, jobID)
				},
			),
			gocron.AfterJobRunsWithError(
				func(jobID uuid.UUID, jobName string, err error) {
					log8.BaseLogger.Debug().Stack().Msgf("The job '%s' (%s) has found errors: %s", jobName, jobID, err)
					log8.BaseLogger.Info().Msgf("The following the GoCron job has found errors: %s(%s).", jobName, jobID)

					// Send notification using the shared notification utility
					message := fmt.Sprintf("Scheduled job '%s' (%s) failed: %v", jobName, jobID, err)
					notificationErr := notification8.Helper.PublishSysErrorNotification(message, "urgent")
					if notificationErr != nil {
						log8.BaseLogger.Error().Err(notificationErr).Msg("Failed to send job error notification")
					}
				},
			),
		)

		// scheduler tags
		BaseReportJobOptionsTags = gocron.WithTags("reportm8")

		BaseReportScheduler.Start()
	})

	return BaseReportScheduler, BaseReportJobOptionsEventListener, BaseReportJobOptionsTags
}
