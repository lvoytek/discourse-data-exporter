package main

import (
	"log"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/lvoytek/discourse_client_go/pkg/discourse"
)

func main() {
	var (
		discourseSiteURL       = kingpin.Flag("discourse.site-url", "The URL of the Discourse site to collect metrics from.").Default("http://127.0.0.1:3000").String()
		discourseCategoryList  = kingpin.Flag("discourse.limit-categories", "Comma separated list of category slugs to limit metrics to. All are enabled by default.").Default("").String()
		discourseRateLimit     = kingpin.Flag("discourse.rate-limit", "Time in seconds to delay each thread's call to Discourse site").Default("5").Int()
		dataCollectOnce        = kingpin.Flag("data.collect-once", "Only collect data once then exit.").Default("false").Bool()
		dataCollectionInterval = kingpin.Flag("data.collection-interval", "Time in seconds to wait before collecting new data from the Discourse site.").Default("3600").Int()
		exportType             = kingpin.Flag("data.export-type", "How to export the data: csv, json, or mysql").Default("mysql").String()
		mysqlServerURL         = kingpin.Flag("mysql.database-url", "The location of the database to export to in mysql mode.").Default("localhost").String()
		mysqlUsername          = kingpin.Flag("mysql.username", "The MySQL user to use for inputting data in mysql mode.").String()
		mysqlPassword          = kingpin.Flag("mysql.password", "The password for the MySQL user to use in mysql mode.").String()
	)

	kingpin.Parse()

	discourseClient := discourse.NewAnonymousClient(*discourseSiteURL)

	exporterErr := InitExporter(*exportType, *mysqlServerURL, *mysqlUsername, *mysqlPassword)

	if exporterErr != nil {
		log.Fatal(exporterErr)
	}

	if *dataCollectOnce {
		discourseData := Collect(discourseClient, strings.Split(strings.TrimSpace(*discourseCategoryList), ","), time.Duration(*discourseRateLimit)*time.Second)
		ExportAll(discourseData, *exportType)
	} else {
		go IntervalCollectAndExport(discourseClient, *exportType, strings.Split(strings.TrimSpace(*discourseCategoryList), ","), time.Duration(*dataCollectionInterval)*time.Second, time.Duration(*discourseRateLimit)*time.Second)
	}
}

func IntervalCollectAndExport(discourseClient *discourse.Client, exportType string, categoryList []string, interval time.Duration, rateLimit time.Duration) {
	for {
		discourseData := Collect(discourseClient, categoryList, rateLimit)
		ExportAll(discourseData, exportType)
		time.Sleep(interval)
	}
}
