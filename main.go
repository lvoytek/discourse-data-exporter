package main

import (
	"log"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/lvoytek/discourse_client_go/pkg/discourse"
)

func main() {
	var (
		discourseSiteURL       = kingpin.Flag("discourse.site-url", "The URL of the Discourse site to collect metrics from.").Default("http://127.0.0.1:3000").String()
		discourseCategory      = kingpin.Flag("discourse.category", "Limit data collected to this category slug.").Default("").String()
		discourseTopic         = kingpin.Flag("discourse.topic", "Limit data collected to this topic ID, overrides discourse.category.").Default("0").Int()
		discourseRateLimit     = kingpin.Flag("discourse.rate-limit", "Time in seconds to delay each thread's call to Discourse site").Default("1").Int()
		dataCollectOnce        = kingpin.Flag("data.collect-once", "Only collect data once then exit.").Default("false").Bool()
		dataCollectionInterval = kingpin.Flag("data.collection-interval", "Time in seconds to wait before collecting new data from the Discourse site.").Default("3600").Int()
		exportType             = kingpin.Flag("data.export-type", "How to export the data: csv, json, or mysql").Default("mysql").String()
		mysqlServerURL         = kingpin.Flag("mysql.database-url", "The location of the database to export to in mysql mode.").Default("localhost").String()
		mysqlUsername          = kingpin.Flag("mysql.username", "The MySQL user to use for inputting data in mysql mode.").String()
		mysqlPassword          = kingpin.Flag("mysql.password", "The password for the MySQL user to use in mysql mode.").String()
		csvFoldername          = kingpin.Flag("csv.foldername", "The name of the folder to send csv files to.").Default("out").String()
		exportTopicComments    = kingpin.Flag("export.posts", "Export posts/comments for each topic.").Default("false").Bool()
		exportTopicEdits       = kingpin.Flag("export.edits", "Export edits to the main post for each topic.").Default("false").Bool()
		exportUsers            = kingpin.Flag("export.users", "Export user metadata").Default("false").Bool()
	)

	kingpin.Parse()

	discourseClient := discourse.NewAnonymousClient(*discourseSiteURL)

	exporterErr := InitExporter(*exportType, *mysqlServerURL, *mysqlUsername, *mysqlPassword, *csvFoldername)

	if exporterErr != nil {
		log.Fatal(exporterErr)
	}

	itemsToExport := ItemsToExport{
		TopicComments: *exportTopicComments,
		TopicEdits:    *exportTopicEdits,
		Users:         *exportUsers,

		LimitToCategorySlug: *discourseCategory,
		LimitToTopicID:      *discourseTopic,
	}

	if *dataCollectOnce {
		discourseData := Collect(discourseClient, itemsToExport, time.Duration(*discourseRateLimit)*time.Second)
		ExportAll(discourseData, *exportType, itemsToExport)
	} else {
		go IntervalCollectAndExport(discourseClient, *exportType, itemsToExport, time.Duration(*dataCollectionInterval)*time.Second, time.Duration(*discourseRateLimit)*time.Second)
	}
}

func IntervalCollectAndExport(discourseClient *discourse.Client, exportType string, itemsToExport ItemsToExport, interval time.Duration, rateLimit time.Duration) {
	for {
		discourseData := Collect(discourseClient, itemsToExport, rateLimit)
		ExportAll(discourseData, exportType, itemsToExport)
		time.Sleep(interval)
	}
}
