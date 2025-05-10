package model8

type TemplateDataEmail struct {
	Monthyear         string
	Totaldomains      int
	Totalhostnames    int
	Totalservices     int
	Totaltechnologies int
	Newhostnames      int
	Nowcritical       int
	Nowhigh           int
	Nowmedium         int
	Nowlow            int
	Nowinfo           int
	Pastcritical      int
	Pasthigh          int
	Pastmedium        int
	Pastlow           int
	Pastinfo          int
	Deltacriticial    int
	Deltahigh         int
	Deltamedium       int
	Deltalow          int
	Deltainfo         int
}

type TemplateDataReport struct {
	Main            TemplateDataReportMain
	Statistics      TemplateDataReportStatistics
	Vulnerabilities []TemplateDataReportVulnerability
	Scope           []TemplateDataScope
}

type TemplateDataReportMain struct {
	Dashboard              string
	Clientorganisation     string
	Reportdate             string
	Securityconsultantname string
	Clientname             string
	Clientemail            string
	Hostname               []TemplateDataReportHostname
	Reportmonth            string
	Reportyear             int
	Securityposture        string
	Totalnumberfindings    int
}

type TemplateDataReportHostname struct {
	Name string
}

type TemplateDataReportStatistics struct {
	Overall   TemplateDataReportOverallStatistics
	Breakdown []TemplateDataReportBreakdownStatistics
}

type TemplateDataReportOverallStatistics struct {
	Reportmonth     string
	Reportpastmonth string
	Criticalnow     int
	Criticalpast    int
	Criticaldiff    int
	Highnow         int
	Highpast        int
	Highdiff        int
	Mediumnow       int
	Mediumpast      int
	Mediumdiff      int
	Lownow          int
	Lowpast         int
	Lowdiff         int
	Infonow         int
	Infopast        int
	Infodiff        int
}

type TemplateDataReportBreakdownStatistics struct {
	Hostname        string
	Reportmonth     string
	Reportpastmonth string
	Criticalnow     int
	Criticalpast    int
	Criticaldiff    int
	Highnow         int
	Highpast        int
	Highdiff        int
	Mediumnow       int
	Mediumpast      int
	Mediumdiff      int
	Lownow          int
	Lowpast         int
	Lowdiff         int
	Infonow         int
	Infopast        int
	Infodiff        int
}

type TemplateDataReportVulnerability struct {
	Hostname      string
	Vulnerability []TemplateDataVulnerabilityDetails
}

type TemplateDataVulnerabilityDetails struct {
	Vulnindex             int
	Vulnrisklevel         string
	Vulnname              string
	Vulnconsequence       string
	Vulnlikelihood        string
	Vulndefinition        string
	Vulnlocation          string
	Vulndescription       string
	Vulnimpact            string
	Vulnreproductionsteps string
	Vulnmitigation        string
}

type TemplateDataScope struct {
	Cell []Hostname8
}
