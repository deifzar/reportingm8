package reporting8

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"deifzar/reportingm8/pkg/cleanup8"
	"deifzar/reportingm8/pkg/cloud8"
	"deifzar/reportingm8/pkg/db8"
	"deifzar/reportingm8/pkg/email8"
	"deifzar/reportingm8/pkg/log8"
	"deifzar/reportingm8/pkg/model8"
	"deifzar/reportingm8/pkg/notification8"
	"deifzar/reportingm8/pkg/utils"
	templatehtml "html/template"
	templatetxt "text/template"

	"github.com/spf13/viper"
)

type Reporting8 struct {
	Db     *sql.DB
	Config *viper.Viper
}

func NewReporting8(db *sql.DB, cnfg *viper.Viper) Reporting8Interface {
	return &Reporting8{Db: db, Config: cnfg}
}

// Emails get delivered the first day of the next month.
func (r *Reporting8) CreateAndSendEmailSummary() error {
	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}
	// Fetch 'user' role type users. If none are found, we triger an error
	DB := r.Db
	user8 := db8.NewDb8User8(DB)
	users, err := user8.GetUsersByRole(model8.RoleUser)
	if err != nil {
		log8.BaseLogger.Error().Msg("CreateAndSendEmailSummary - errors triggered when fetching `user` role type users")
		notification8.Helper.PublishSysErrorNotification("CreateAndSendEmailSummary - errors triggered when fetching `user` role type users", "urgent", "reportingm8")
		return err
	}
	if len(users) < 1 {
		log8.BaseLogger.Warn().Msg("CreateAndSendEmailSummary - at least one customer must be enrolled in the system before sending any email summary")
		notification8.Helper.PublishSysWarningNotification("CreateAndSendEmailSummary - at least one customer must be enrolled in the system before sending any email summary", "normal", "reportingm8")
		return errors.New("CreateAndSendEmailSummary (Warning) - at least one customer must be enrolled in the system before sending any email summary")
	}
	// create email summary
	message, err := r.getEmailSummaryMessage()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		notification8.Helper.PublishSysErrorNotification("CreateAndSendEmailSummary - generating email summary message", "urgent", "reportingm8")
		return err
	}

	// send email summary notification
	email8, err := email8.NewEmail8()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("CreateAndSendEmailSummary - errors triggered when initializing the email settings")
		notification8.Helper.PublishSysErrorNotification("CreateAndSendEmailSummary - errors triggered when initializing the email settings", "urgent", "reportingm8")
		return err
	}
	err = email8.SendEmail([]string{"no-reply@cptm8.net", "info@deifzar.me"}, message.Bytes())
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("CreateAndSendEmailSummary - errors triggered when sending the Email html file")
		notification8.Helper.PublishSysErrorNotification("CreateAndSendEmailSummary - errors triggered when sending the Email html file", "urgent", "reportingm8")
		return err
	}
	return nil
}

// Reports are delivered by default the first day of the next month.
func (r *Reporting8) CreateReportAndSendEmailNotification(companyName string) error {
	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}
	// Fetch 'user' role type users. If none are found, we triger an error
	DB := r.Db
	user8 := db8.NewDb8User8(DB)
	users, err := user8.GetUsersByRole(model8.RoleUser)
	if err != nil {
		log8.BaseLogger.Error().Msg("CreateReportAndSendEmailNotification - errors triggered when fetching `user` role type users")
		notification8.Helper.PublishSysErrorNotification("CreateReportAndSendEmailNotification - errors triggered when fetching `user` role type users", "urgent", "reportingm8")
		return err
	}
	if len(users) < 1 || !user8.ExistReportAuthor() || !user8.ExistReportOwner() {
		log8.BaseLogger.Warn().Msg("CreateReportAndSendEmailNotification - customers, one report author and one report owner must be set in the system before sending any report")
		notification8.Helper.PublishSysWarningNotification("CreateReportAndSendEmailNotification - customers, one report author and one report owner must be set in the system before sending any report", "normal", "reportingm8")
		return errors.New("CreateReportAndSendEmailNotification (Warning) - customers, one report author and one report owner must be set in the system before sending any report")
	}
	// crete report and return data template
	datareport, reportfilename, err := r.createReport(companyName)
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		log8.BaseLogger.Debug().Msg(err.Error())
		notification8.Helper.PublishSysErrorNotification("CreateReportAndSendEmailNotification - creating report", "urgent", "reportingm8")
		return err
	}

	// Upload Report to the cloud repo - AWS S3
	repoURI, err := r.uploadReportIntoRepo(reportfilename)
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		notification8.Helper.PublishSysErrorNotification("CreateReportAndSendEmailNotification - uploading report to cloud repo", "urgent", "reportingm8")
		return err
	}
	// add report stats and download link into DB
	err = r.addDataReportIntoDB(datareport, repoURI)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		notification8.Helper.PublishSysErrorNotification("CreateReportAndSendEmailNotification - inserting report in database", "urgent", "reportingm8")
		return err
	}

	message, err := r.getEmailReportNotificationMessage(repoURI)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		notification8.Helper.PublishSysErrorNotification("CreateReportAndSendEmailNotification - creating email template", "urgent", "reportingm8")
		return err
	}
	// Send email to all `user` role type users informing that a new report is ready in the dashboard
	email8, err := email8.NewEmail8()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("CreateReportAndSendEmailNotification - errors triggered when initializing the email settings")
		notification8.Helper.PublishSysErrorNotification("CreateReportAndSendEmailNotification - errors triggered when initializing the email settings", "urgent", "reportingm8")
		return err
	}
	// Question: Send email to all `user` role type users or only with the 'report' flag enabled?
	// TODO: fetch users
	err = email8.SendEmail([]string{"no-reply@cptm8.net", "info@deifzar.me"}, message.Bytes())
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("CreateReportAndSendEmailNotification - errors triggered when sending the Email html file")
		notification8.Helper.PublishSysErrorNotification("CreateReportAndSendEmailNotification - ", "urgent", "reportingm8")
		return err
	}

	return nil
}

func (r *Reporting8) getEmailSummaryMessage() (bytes.Buffer, error) {
	var body bytes.Buffer

	// email Template file
	var templatefile = r.Config.GetString("REPORTINGM8.Template.emailsummary")
	thtml, err := templatehtml.ParseFiles(templatefile)
	if err != nil {
		log8.BaseLogger.Error().Msg("errors triggered when parsing the Email summary html template file")
		return bytes.Buffer{}, err
	}

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	// // Previous month
	// prevmonthdate := time.Now()
	// year, month, _ := prevmonthdate.Date()
	// monthyear := month.String() + " " + strconv.Itoa(year)
	body.Write([]byte(fmt.Sprintf("Subject: CPT High level Summary\n%s\n\n", headers)))

	templatedata, err := r.getTemplateDataSummary()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("errors triggered when fetching data for the Email summary html template  file")
		return bytes.Buffer{}, err
	}
	err = thtml.Execute(&body, templatedata)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("errors triggered when parsing data into the Email summary html template file")
		return bytes.Buffer{}, err
	}
	return body, nil
}

func (r *Reporting8) getEmailReportNotificationMessage(link string) (bytes.Buffer, error) {
	var body bytes.Buffer
	var emailnotificationreporttemplatefile = r.Config.GetString("REPORTINGM8.Template.emailnotificationreport")
	thtml, err := templatehtml.ParseFiles(emailnotificationreporttemplatefile)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("errors triggered when parsing the Email Report notification html template file")
		return bytes.Buffer{}, err
	}
	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	// prevmonthdate := time.Now().AddDate(0, -1, 0)
	// year, month, _ := prevmonthdate.Date()
	// monthyear := month.String() + " " + strconv.Itoa(year)
	body.Write([]byte(fmt.Sprintf("Subject: CPT New Report Available\n%s\n\n", headers)))

	err = thtml.Execute(&body, struct {
		ReportDashboardLink string
	}{
		ReportDashboardLink: link,
	})
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("errors triggered when parsing data into the Email summary html template file")
		return bytes.Buffer{}, err
	}
	return body, nil

}

func (r *Reporting8) createReport(companyName string) (model8.TemplateDataReport, string, error) {
	// report Template files
	var file *os.File
	var reporttemplatefile = r.Config.GetString("REPORTINGM8.Template.report")

	tmpl, err := templatetxt.ParseFiles(reporttemplatefile)
	if err != nil {
		log8.BaseLogger.Error().Msg("errors triggered when parsing the template Report HTML file")
		return model8.TemplateDataReport{}, "", err
	}
	// fetch temp data for report
	data, err := r.getTemplateDataReport(companyName)
	if err != nil {
		log8.BaseLogger.Error().Msg("errors triggered when fetching data for the template Report HTML file")
		return model8.TemplateDataReport{}, "", err
	}
	// Report date
	reportdate := time.Now() // .AddDate(0, -1, 0)
	tempfile := "tmp/output/" + reportdate.Format("2006-01-02") + "-report.html"
	file, err = os.Create(tempfile)
	if err != nil {
		log8.BaseLogger.Error().Msg("errors triggered when creating the temp Report html report file")
		return model8.TemplateDataReport{}, "", err
	}
	defer file.Close()
	err = tmpl.Execute(file, data)
	if err != nil {
		log8.BaseLogger.Error().Msg("errors triggered when parsing data into the Report html template file")
		return model8.TemplateDataReport{}, "", err
	}
	return data, tempfile, nil
}

func (r *Reporting8) uploadReportIntoRepo(filename string) (string, error) {
	c8, err := cloud8.NewCloud8()
	if err != nil {
		log8.BaseLogger.Error().Msg("errors triggered when initialising cloud interface")
		return "", err
	}
	uri, err := c8.UploadToBucket(filename)
	if err != nil {
		log8.BaseLogger.Error().Msg("errors triggered when upload the file into the cloud")
		return "", err
	}
	return uri, nil
}

// Add cptm8 report new entry into database
func (r *Reporting8) addDataReportIntoDB(data model8.TemplateDataReport, link string) error {
	DB := r.Db
	report8DB := db8.NewDbReport8(DB)
	datasummary, err := r.getTemplateDataSummary()
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		log8.BaseLogger.Error().Msg("sendReport - errors triggered when fetching data from the DB")
		return err
	}
	_, err = report8DB.InsertReport(
		model8.Report8{
			Nvul:      data.Main.Totalnumberfindings,
			Ndomain:   datasummary.Totaldomains,
			Nhostname: datasummary.Totalhostnames,
			Ncritical: datasummary.Nowcritical,
			Nhigh:     datasummary.Nowhigh,
			Nmedium:   datasummary.Nowmedium,
			Nlow:      datasummary.Nowlow,
			Ninfo:     datasummary.Nowinfo,
			Nservice:  datasummary.Totalservices,
			Ntech:     datasummary.Totaltechnologies,
			Release:   time.Now().AddDate(0, -1, 0),
			Link:      link,
		},
	)
	if err != nil {
		log8.BaseLogger.Error().Msg("sendReport - errors triggered when inserting data into the report DB table")
		return err
	}
	return nil
}

func (r *Reporting8) getTemplateDataScope() ([]model8.TemplateDataScope, error) {
	DB := r.Db
	hostname8 := db8.NewDb8Hostname8(DB)
	hostnames, err := hostname8.GetAllEnabled()
	if err != nil {
		return nil, err
	}
	var template []model8.TemplateDataScope
	limit := 3
	for i := 0; i < len(hostnames); i += limit {
		r := model8.TemplateDataScope{
			Cell: hostnames[i:min(i+limit, len(hostnames))],
		}
		template = append(template, r)
	}

	return template, err
}

func (r *Reporting8) getTemplateDataVulnerability() ([]model8.TemplateDataReportVulnerability, error) {
	DB := r.Db
	hostname8 := db8.NewDb8Hostname8(DB)
	vulnerability8 := db8.NewDb8Vulnerability8(DB)
	hostnames, err := hostname8.Get()
	if err != nil {
		return nil, err
	}
	var template []model8.TemplateDataReportVulnerability
	if len(hostnames) > 0 {
		vulnindex := 1
		for _, h := range hostnames {
			vulnerabilities, err := vulnerability8.GetAllNotResolvedByHostnameId(h.Id)
			if err != nil {
				log8.BaseLogger.Info().Msgf("error returned when fetching all `not resolved` findings for the hostname: %s", h.Name)
				log8.BaseLogger.Info().Msg("continue to the next hostname")
				continue
			}
			if len(vulnerabilities) > 0 {
				var listvulndetails []model8.TemplateDataVulnerabilityDetails
				for _, vuln := range vulnerabilities {
					var definition, location, description, impact, reproductionsteps, mitigation string
					if vuln.Definitionhtml.Valid {
						definition = vuln.Definitionhtml.String
					} else {
						definition = ""
					}
					if vuln.Locationhtml.Valid {
						location = vuln.Locationhtml.String
					} else {
						location = ""
					}
					if vuln.Descriptionhtml.Valid {
						description = vuln.Descriptionhtml.String
					} else {
						description = ""
					}
					if vuln.Impacthtml.Valid {
						impact = vuln.Impacthtml.String
					} else {
						impact = ""
					}
					if vuln.Reproductionstepshtml.Valid {
						reproductionsteps = vuln.Reproductionstepshtml.String
					} else {
						reproductionsteps = ""
					}
					if vuln.Mitigationhtml.Valid {
						mitigation = vuln.Mitigationhtml.String
					} else {
						mitigation = ""
					}
					vulndetails := model8.TemplateDataVulnerabilityDetails{
						Vulnindex:             vulnindex,
						Vulnname:              vuln.Title,
						Vulnrisklevel:         vuln.Risklevelname,
						Vulnconsequence:       vuln.Riskconsequencename,
						Vulnlikelihood:        vuln.Risklikelihoodname,
						Vulndefinition:        definition,
						Vulnlocation:          location,
						Vulndescription:       description,
						Vulnimpact:            impact,
						Vulnreproductionsteps: reproductionsteps,
						Vulnmitigation:        mitigation,
					}
					listvulndetails = append(listvulndetails, vulndetails)
					vulnindex = vulnindex + 1
				}
				item := model8.TemplateDataReportVulnerability{
					Hostname:      h.Name,
					Vulnerability: listvulndetails,
				}
				template = append(template, item)
			}
		}
	}
	return template, nil
}

func (r *Reporting8) getTemplateDataReportBreakdownStatistics() ([]model8.TemplateDataReportBreakdownStatistics, error) {
	DB := r.Db
	vulnerability8 := db8.NewDb8Vulnerability8(DB)
	statsnow, err := vulnerability8.GetStatsLastMonth()
	if err != nil {
		return nil, err
	}
	statspast, err := vulnerability8.GetStatsLastTwoMonths()
	if err != nil {
		return nil, err
	}
	stats := utils.MergeMaps(statsnow, statspast)

	// Report stats dates
	reportdate := time.Now().AddDate(0, -1, 0)
	reportpastdate := time.Now().AddDate(0, -2, 0)
	year, month, _ := reportdate.Date()
	pastyear, pastmonth, _ := reportpastdate.Date()
	reportmonth := month.String() + " " + strconv.Itoa(year)
	reportpastmonth := pastmonth.String() + " " + strconv.Itoa(pastyear)
	var template []model8.TemplateDataReportBreakdownStatistics
	for key, stat := range stats {
		var item = model8.TemplateDataReportBreakdownStatistics{
			Hostname:        key,
			Reportmonth:     reportmonth,
			Reportpastmonth: reportpastmonth,
			Criticalnow:     stat["pastonecritical"],
			Criticalpast:    stat["pasttwocritical"],
			Criticaldiff:    stat["pastonecritical"] - stat["pasttwocritical"],
			Highnow:         stat["pastonehigh"],
			Highpast:        stat["pasttwohigh"],
			Highdiff:        stat["pastonehigh"] - stat["pasttwohigh"],
			Mediumnow:       stat["pastonemedium"],
			Mediumpast:      stat["pasttwomedium"],
			Mediumdiff:      stat["pastonemedium"] - stat["pasttwomedium"],
			Lownow:          stat["pastonelow"],
			Lowpast:         stat["pasttwolow"],
			Lowdiff:         stat["pastonelow"] - stat["pasttwolow"],
			Infonow:         stat["pastoneinfo"],
			Infopast:        stat["pasttwoinfo"],
			Infodiff:        stat["pastoneinfo"] - stat["pasttwoinfo"],
		}
		template = append(template, item)
	}
	return template, nil
}

func (r *Reporting8) getTemplateDataReportOverallStatistics() (model8.TemplateDataReportOverallStatistics, error) {
	DB := r.Db
	vulnerability8 := db8.NewDb8Vulnerability8(DB)
	nowvulnstotals, err := vulnerability8.CountTotalsLastMonth()
	if err != nil {
		return model8.TemplateDataReportOverallStatistics{}, err
	}
	pastvulnstotals, err := vulnerability8.CountTotalsLastTwoMonths()
	if err != nil {
		return model8.TemplateDataReportOverallStatistics{}, err
	}
	// Report stats dates
	reportdate := time.Now().AddDate(0, -1, 0)
	reportpastdate := time.Now().AddDate(0, -2, 0)
	year, month, _ := reportdate.Date()
	pastyear, pastmonth, _ := reportpastdate.Date()
	reportmonth := month.String() + " " + strconv.Itoa(year)
	reportpastmonth := pastmonth.String() + " " + strconv.Itoa(pastyear)
	template := model8.TemplateDataReportOverallStatistics{
		Reportmonth:     reportmonth,
		Reportpastmonth: reportpastmonth,
		Criticalnow:     nowvulnstotals["critical"],
		Criticalpast:    pastvulnstotals["critical"],
		Criticaldiff:    pastvulnstotals["critical"] - nowvulnstotals["critical"],
		Highnow:         nowvulnstotals["high"],
		Highpast:        pastvulnstotals["high"],
		Highdiff:        pastvulnstotals["high"] - nowvulnstotals["high"],
		Mediumnow:       nowvulnstotals["medium"],
		Mediumpast:      pastvulnstotals["medium"],
		Mediumdiff:      pastvulnstotals["medium"] - nowvulnstotals["medium"],
		Lownow:          nowvulnstotals["low"],
		Lowpast:         pastvulnstotals["low"],
		Lowdiff:         pastvulnstotals["low"] - nowvulnstotals["low"],
		Infonow:         nowvulnstotals["info"],
		Infopast:        pastvulnstotals["info"],
		Infodiff:        pastvulnstotals["info"] - nowvulnstotals["info"],
	}
	return template, nil
}

func (r *Reporting8) getTemplateDataReportStatistics() (model8.TemplateDataReportStatistics, error) {
	overall, err := r.getTemplateDataReportOverallStatistics()
	if err != nil {
		return model8.TemplateDataReportStatistics{}, err
	}
	breakdown, err := r.getTemplateDataReportBreakdownStatistics()
	if err != nil {
		return model8.TemplateDataReportStatistics{}, err
	}
	template := model8.TemplateDataReportStatistics{
		Overall:   overall,
		Breakdown: breakdown,
	}
	return template, nil
}

func (r *Reporting8) getTemplateDataReportMain(companyNameReport string) (model8.TemplateDataReportMain, error) {
	DB := r.Db
	vulnerability8 := db8.NewDb8Vulnerability8(DB)
	user8 := db8.NewDb8User8(DB)
	// Get hostnames only with vulnerabilities
	stats, err := vulnerability8.GetStatsLastMonth()
	if err != nil {
		return model8.TemplateDataReportMain{}, err
	}
	var templatehostanme []model8.TemplateDataReportHostname
	for k := range stats {
		templatehostanme = append(templatehostanme, model8.TemplateDataReportHostname{Name: k})
	}
	nowvulnstotals, err := vulnerability8.CountTotalsLastMonth()
	if err != nil {
		return model8.TemplateDataReportMain{}, err
	}
	reportowner, err := user8.GetReportOwner()
	if err != nil {
		return model8.TemplateDataReportMain{}, err
	}
	if (reportowner == model8.User8{}) {
		return model8.TemplateDataReportMain{}, errors.New("report owner not found")
	}
	reportauthor, err := user8.GetReportAuthor()
	if err != nil {
		return model8.TemplateDataReportMain{}, err
	}
	if (reportauthor == model8.User8{}) {
		return model8.TemplateDataReportMain{}, errors.New("report author not found")
	}
	securityposture := utils.GetSecurityPosture(nowvulnstotals)

	// Report date
	reportdate := time.Now().AddDate(0, -1, 0)
	year, month, _ := reportdate.Date()
	template := model8.TemplateDataReportMain{
		Clientorganisation:     companyNameReport,
		Dashboard:              r.Config.GetString("REPORTINGM8.Template.dashboard"),
		Reportdate:             reportdate.Format("02/01/2006"),
		Securityconsultantname: reportauthor.Name,
		Clientname:             reportowner.Name,
		Clientemail:            reportowner.Email,
		Hostname:               templatehostanme,
		Reportmonth:            month.String(),
		Reportyear:             year,
		Securityposture:        securityposture,
		Totalnumberfindings:    nowvulnstotals["nvul"],
	}
	return template, nil
}

func (r *Reporting8) getTemplateDataReport(companyName string) (model8.TemplateDataReport, error) {
	main, err := r.getTemplateDataReportMain(companyName)
	if err != nil {
		return model8.TemplateDataReport{}, err
	}
	statistics, err := r.getTemplateDataReportStatistics()
	if err != nil {
		return model8.TemplateDataReport{}, err
	}
	vulnerabilities, err := r.getTemplateDataVulnerability()
	if err != nil {
		return model8.TemplateDataReport{}, err
	}
	scope, err := r.getTemplateDataScope()
	if err != nil {
		return model8.TemplateDataReport{}, err
	}
	template := model8.TemplateDataReport{
		Main:            main,
		Statistics:      statistics,
		Vulnerabilities: vulnerabilities,
		Scope:           scope,
	}
	return template, nil
}

func (r *Reporting8) getTemplateDataSummary() (model8.TemplateDataEmail, error) {
	DB := r.Db
	domain8 := db8.NewDb8Domain8(DB)
	hostname8 := db8.NewDb8Hostname8(DB)
	hostnameinfo8 := db8.NewDb8Hostnameinfo8(DB)
	vulnerability8 := db8.NewDb8Vulnerability8(DB)
	service8 := db8.NewDb8Service8(DB)
	totaldomains, err := domain8.Count(true)
	if err != nil {
		return model8.TemplateDataEmail{}, err
	}
	totalhostnames, err := hostname8.Count(true)
	if err != nil {
		return model8.TemplateDataEmail{}, err
	}
	totalsoftware, err := hostnameinfo8.CountUniqueSoftware()
	if err != nil {
		return model8.TemplateDataEmail{}, err
	}
	newhostnames, err := hostname8.CountFoundLastMonth()
	if err != nil {
		return model8.TemplateDataEmail{}, err
	}
	totalservices, err := service8.CountUniqueServices()
	if err != nil {
		return model8.TemplateDataEmail{}, err
	}
	nowvulnstotals, err := vulnerability8.CountTotalsLastMonth()
	if err != nil {
		return model8.TemplateDataEmail{}, err
	}
	pastvulnstotals, err := vulnerability8.CountTotalsLastTwoMonths()
	if err != nil {
		return model8.TemplateDataEmail{}, err
	}
	// Report date
	reportdate := time.Now().AddDate(0, -1, 0)
	year, month, _ := reportdate.Date()
	monthyear := month.String() + " " + strconv.Itoa(year)

	templatedataEmail := model8.TemplateDataEmail{
		Monthyear:         monthyear,
		Totaldomains:      totaldomains,
		Totalhostnames:    totalhostnames,
		Totalservices:     totalservices,
		Totaltechnologies: totalsoftware,
		Newhostnames:      newhostnames,
		Nowcritical:       nowvulnstotals["critical"],
		Nowhigh:           nowvulnstotals["high"],
		Nowmedium:         nowvulnstotals["medium"],
		Nowlow:            nowvulnstotals["low"],
		Nowinfo:           nowvulnstotals["info"],
		Pastcritical:      pastvulnstotals["critical"],
		Pasthigh:          pastvulnstotals["high"],
		Pastmedium:        pastvulnstotals["medium"],
		Pastlow:           pastvulnstotals["low"],
		Pastinfo:          pastvulnstotals["info"],
		Deltacriticial:    pastvulnstotals["critical"] - nowvulnstotals["critical"],
		Deltahigh:         pastvulnstotals["high"] - nowvulnstotals["high"],
		Deltamedium:       pastvulnstotals["medium"] - nowvulnstotals["medium"],
		Deltalow:          pastvulnstotals["low"] - nowvulnstotals["low"],
		Deltainfo:         pastvulnstotals["info"] - nowvulnstotals["info"],
	}
	return templatedataEmail, nil
}
