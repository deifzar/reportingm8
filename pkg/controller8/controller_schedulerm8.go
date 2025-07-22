package controller8

import (
	"database/sql"
	"net/http"
	"time"

	"deifzar/reportingm8/pkg/db8"
	"deifzar/reportingm8/pkg/gocron8"
	"deifzar/reportingm8/pkg/log8"
	"deifzar/reportingm8/pkg/model8"
	"deifzar/reportingm8/pkg/reporting8"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"

	"github.com/spf13/viper"
)

type Schedulerm8 struct {
	Db     *sql.DB
	Config *viper.Viper
}

func NewSchedulerm8(db *sql.DB, cnfg *viper.Viper) Scheduler8Interface {
	return &Schedulerm8{Db: db, Config: cnfg}
}

func (s *Schedulerm8) InitScheduler() error {
	// Remove all gocron jobs and init new ones
	gocron8.BaseReportScheduler.RemoveByTags("reportm8")

	r8 := reporting8.NewReporting8(s.Db, s.Config)
	rss8 := db8.NewDb8ReportSchedulingSettings8(s.Db)
	reportSettings8, err := rss8.Get()
	if err != nil {
		return err
	}
	var emailjob, reportjob gocron.JobDefinition
	var companyNameReport string = s.Config.GetString("REPORTINGM8.Template.clientorganisation")
	if reportSettings8 == (model8.Reportschedulingsettings8{}) {
		log8.BaseLogger.Info().Msg("InitScheduler - Empty reportschedulingsettings DB table. Creating email and report jobs with default settings")
		// emailjon default weekly
		emailjob = gocron.WeeklyJob(
			1,
			gocron.NewWeekdays(time.Monday),
			gocron.NewAtTimes(
				gocron.NewAtTime(10, 0, 0),
			),
		)
		// reportjob default monthly
		reportjob = gocron.MonthlyJob(
			1,
			gocron.NewDaysOfTheMonth(1),
			gocron.NewAtTimes(
				gocron.NewAtTime(10, 0, 0),
			),
		)
	} else {
		log8.BaseLogger.Info().Msg("InitScheduler - No Empty reportschedulingsettings DB table. Creating email and report jobs with saved settings")
		// email scheduling
		switch reportSettings8.Settings.Email.EmailTimebase {
		case "weekly":
			emailjob = gocron.WeeklyJob(
				uint(reportSettings8.Settings.Email.EmailFrequency),
				gocron.NewWeekdays(time.Monday),
				gocron.NewAtTimes(
					gocron.NewAtTime(10, 0, 0),
				),
			)
		case "monthly":
			emailjob = gocron.MonthlyJob(
				uint(reportSettings8.Settings.Email.EmailFrequency),
				gocron.NewDaysOfTheMonth(1),
				gocron.NewAtTimes(
					gocron.NewAtTime(10, 0, 0),
				),
			)
		default:
			emailjob = gocron.MonthlyJob(
				1,
				gocron.NewDaysOfTheMonth(1),
				gocron.NewAtTimes(
					gocron.NewAtTime(10, 0, 0),
				),
			)
		}
		// document scheduling
		switch reportSettings8.Settings.Document.DocumentTimebase {
		case "weekly":
			reportjob = gocron.WeeklyJob(
				uint(reportSettings8.Settings.Document.DocumentFrequency),
				gocron.NewWeekdays(time.Monday),
				gocron.NewAtTimes(
					gocron.NewAtTime(10, 0, 0),
				),
			)
		case "monthly":
			reportjob = gocron.MonthlyJob(
				uint(reportSettings8.Settings.Document.DocumentFrequency),
				gocron.NewDaysOfTheMonth(1),
				gocron.NewAtTimes(
					gocron.NewAtTime(10, 0, 0),
				),
			)
		default:
			reportjob = gocron.MonthlyJob(
				1,
				gocron.NewDaysOfTheMonth(1),
				gocron.NewAtTimes(
					gocron.NewAtTime(10, 0, 0),
				),
			)
		}
		// company name report
		if reportSettings8.Settings.CompanyNameReport != "" {
			companyNameReport = reportSettings8.Settings.CompanyNameReport
		}
	}
	log8.BaseLogger.Info().Msgf("InitScheduler - The report contains the following company name: '%s'", companyNameReport)
	var emailtask = gocron.NewTask(
		r8.CreateAndSendEmailSummary,
	)
	var reporttask = gocron.NewTask(
		r8.CreateReportAndSendEmailNotification, companyNameReport,
	)

	// Job for email highlevel summary findings
	job, err := gocron8.BaseReportScheduler.NewJob(emailjob, emailtask, gocron.WithName("Email scheduler Job"), gocron8.BaseReportJobOptionsTags, gocron8.BaseReportJobOptionsEventListener, gocron8.BaseReportJobOptionsEventListener)
	if err != nil {
		return err
	}
	log8.BaseLogger.Info().Msgf("InitScheduler - Email Job with name '%s' and ID '%s' was successfully created", job.Name(), job.ID())

	// Job for report and email notification
	job, err = gocron8.BaseReportScheduler.NewJob(reportjob, reporttask, gocron.WithName("Document scheduler Job"), gocron8.BaseReportJobOptionsTags, gocron8.BaseReportJobOptionsEventListener, gocron8.BaseReportJobOptionsEventListener)
	if err != nil {
		return err
	}
	log8.BaseLogger.Info().Msgf("InitScheduler - Report Job with name '%s' and ID '%s' was successfully created", job.Name(), job.ID())
	return nil
}

func (s *Schedulerm8) UpdateScheduler(c *gin.Context) {
	var post model8.PostReportschedulingsettings8
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "UpdateScheduler failed - Check Body parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - UpdateScheduler")
		return
	}
	rss8 := db8.NewDb8ReportSchedulingSettings8(s.Db)
	update, err := rss8.Set(post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "msg": "UpdateScheduler failed due to DB errors."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - UpdateScheduler failed after attempt to update the report settings in the DB")
		return
	}
	err = s.InitScheduler()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "msg": "UpdateScheduler failed due to cronjobs errors."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - UpdateScheduler failed after attempt to set the new cron jobs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": update, "msg": "update report scheduling settings - success"})
	log8.BaseLogger.Info().Msg("200 HTTP Response - UpdateScheduler")
}

func (s *Schedulerm8) GetSchedulerDetails(c *gin.Context) {
	jobs := gocron8.BaseReportScheduler.Jobs()
	if len(jobs) < 1 {
		c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "GetSchedulerDetails returns empty number of jobs."})
		log8.BaseLogger.Info().Msg("200 HTTP Response - GetSchedulerDetails returns an empty number of jobs details")
		return
	}
	var details []model8.Scheduler8
	for _, j := range jobs {
		nextrun, err := j.NextRun()
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			log8.BaseLogger.Warn().Msgf("GetSchedulerDetails - failed when getting the NextRun time for the Job %s (%s)", j.Name(), j.ID())
			continue
		}

		details = append(details, model8.Scheduler8{Id: j.ID(), Name: j.Name(), Tags: j.Tags(), Nextrun: nextrun.String()})
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": details, "msg": "new report scheduling settings - success"})
	log8.BaseLogger.Info().Msg("200 HTTP Response - GetSchedulerDetails")
}
