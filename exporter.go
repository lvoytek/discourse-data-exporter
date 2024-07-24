package main

import (
	"fmt"

	"github.com/lvoytek/discourse_client_go/pkg/discourse"
)

func InitExporter(exportType string, mysqlServerURL string, mysqlUsername string, mysqlPassword string) error {
	if exportType == "mysql" {
		return ConnectMySQL(mysqlServerURL, mysqlUsername, mysqlPassword)
	}

	return fmt.Errorf("Invalid Exporter Type: %s", exportType)
}

func ExportAll(cache DiscourseCache, exportType string) {
	ExportTopicComments(cache.Topics, exportType)
}

func ExportTopicComments(topics map[string]map[int]*discourse.TopicData, exportType string) {
	topicComments := topicMapToTopicComments(topics)

	if exportType == "mysql" {
		ExportTopicCommentsMySQL(topicComments)
	}
}

func topicMapToTopicComments(topics map[string]map[int]*discourse.TopicData) (topicComments []TopicCommentsEntry) {
	for category_slug, topic_list := range topics {
		for topic_id, topic := range topic_list {
			for _, post := range topic.PostStream.Posts {
				topicComments = append(topicComments, TopicCommentsEntry{
					category_slug: category_slug,
					topic_id:      topic_id,
					post_id:       post.ID,
					creation_time: post.CreatedAt,
					update_time:   post.UpdatedAt,
					username:      post.Username,
				})
			}
		}
	}

	return topicComments
}
