package main

import (
	"fmt"

	"github.com/lvoytek/discourse_client_go/pkg/discourse"
)

func InitExporter(exportType string, mysqlServerURL string, mysqlUsername string, mysqlPassword string) error {
	if exportType == "mysql" {
		err := ConnectMySQL(mysqlServerURL, mysqlUsername, mysqlPassword)

		if err != nil {
			return err
		}

		return InitializeMySQLDatabase()
	}

	return fmt.Errorf("invalid exporter type: %s", exportType)
}

func ExportAll(cache DiscourseCache, exportType string) {
	ExportUsers(cache.Topics, exportType)
	ExportTopicComments(cache.Topics, exportType)
}

func ExportUsers(topics map[string]map[int]*discourse.TopicData, exportType string) {
	userEntries := userMapToUserEntry(topics)

	if exportType == "mysql" {
		ExportUsersMySQL(userEntries)
	}
}

func ExportTopicComments(topics map[string]map[int]*discourse.TopicData, exportType string) {
	topicComments := topicMapToTopicComments(topics)

	if exportType == "mysql" {
		ExportTopicCommentsMySQL(topicComments)
	}
}

func userMapToUserEntry(topics map[string]map[int]*discourse.TopicData) (userEntries []UserEntry) {
	for _, topic_list := range topics {
		for _, topic := range topic_list {
			for _, participant := range topic.Details.Participants {
				userEntries = append(userEntries, UserEntry{
					user_id:            participant.ID,
					username:           participant.Username,
					name:               participant.Name,
					primary_group_name: participant.PrimaryGroupName,
				})
			}
		}
	}

	return userEntries
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
