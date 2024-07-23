package main

import (
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/lvoytek/discourse_client_go/pkg/discourse"
)

func main() {
	var (
		discourseSiteURL       = kingpin.Flag("discourse.site-url", "The URL of the Discourse site to collect metrics from.").Default("http://127.0.0.1:3000").String()
		discourseCategoryList  = kingpin.Flag("discourse.limit-categories", "Comma separated list of category slugs to limit metrics to. All are enabled by default.").Default("").String()
		dataCollectOnce        = kingpin.Flag("data.collect-once", "Only collect data once then exit.").Default(true).Boolean()
		dataCollectionInterval = kingpin.Flag("data.collection-interval", "Time in seconds to wait before collecting new data from the Discourse site.").Default("3600").Int()
		exportType             = kingpin.Flag("data.export-type", "How to export the data: csv, json, or mysql").Default("csv").String()
		mysqlServerURL         = kingpin.Flag("mysql.database-url", "The location of the database to export to in mysql mode.").Default("localhost:3306").String()
		mysqlUsername          = kingpin.Flag("mysql.username", "The MySQL user to use for inputting data in mysql mode.").String()
		mysqlPassword          = kingpin.Flag("mysql.password", "The password for the MySQL user to use in mysql mode.").String()
	)

	kingpin.Parse()

	discourseClient := discourse.NewAnonymousClient(*discourseSiteURL)

	exporterErr := InitExporter(*exportType, *mysqlServerURL, *mysqlUsername, *mysqlPassword)

	if exporterErr != nil {
		panic(exporterErr)
	}

	if *dataCollectOnce {
		discourseData := Collect(discourseClient, strings.Split(strings.TrimSpace(*discourseCategoryList), ","))
		ExportAll(discourseData, *exportType)
	} else {
		go IntervalCollectAndExport(discourseClient, *exportType, strings.Split(strings.TrimSpace(*discourseCategoryList), ","), time.Duration(*dataCollectionInterval)*time.Second)
	}
}

func IntervalCollectAndExport(discourseClient *discourse.Client, exportType string, categoryList []string, interval time.Duration) {
	for {
		discourseData := Collect(discourseClient, categoryList)
		ExportAll(discourseData, exportType)
		time.Sleep(interval)
	}
}
