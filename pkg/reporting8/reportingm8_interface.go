package reporting8

import (
	"bytes"
	"deifzar/reportingm8/pkg/model8"
)

type Reporting8Interface interface {
	// Emails get delivered the first day of the next month.
	CreateAndSendEmailSummary() error
	// Reports are delivered by default the first day of the next month.
	CreateReportAndSendEmailNotification(companyName string) error
	// return email summary body message
	getEmailSummaryMessage() (bytes.Buffer, error)
	// return email report notification body message
	getEmailReportNotificationMessage(link string) (bytes.Buffer, error)
	// create report and return data template along with the report full path
	createReport(companyName string) (model8.TemplateDataReport, string, error)
	// return the data required to create the email summary
	getTemplateDataSummary() (model8.TemplateDataEmail, error)
	// return the all the data required to create the final report
	getTemplateDataReport(companyName string) (model8.TemplateDataReport, error)
	// return the data that fits in the first sections of the report: First Page, Document Control, Table of Content, Executive Summary and Service Summary
	getTemplateDataReportMain(companyName string) (model8.TemplateDataReportMain, error)
	// return the data that fits in the Appendix Scope section
	getTemplateDataScope() ([]model8.TemplateDataScope, error)
	// return the data that fits 'Security Findings Breakdown by System' section
	getTemplateDataVulnerability() ([]model8.TemplateDataReportVulnerability, error)
	// return the data that fits in the 'Statistics Breakdown by System' section
	getTemplateDataReportBreakdownStatistics() ([]model8.TemplateDataReportBreakdownStatistics, error)
	// return the data that fits in the 'Monthly Overall Statistics' section
	getTemplateDataReportOverallStatistics() (model8.TemplateDataReportOverallStatistics, error)
	// return the data that fits in the Statistic section. It calls 'getTemplateDataReportOverallStatistics' and 'getTemplateDataReportBreakdownStatistics'
	getTemplateDataReportStatistics() (model8.TemplateDataReportStatistics, error)
	// add data template stats and download link into database
	addDataReportIntoDB(data model8.TemplateDataReport, link string) error
	// upload filename (full path) into the S3 bucket and return URL link
	uploadReportIntoRepo(filename string) (string, error)
}
