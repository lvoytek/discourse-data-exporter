package main

import (
	"fmt"

	"github.com/lvoytek/discourse_client_go/pkg/discourse"
)

func InitExporter(exportType string, mysqlServerURL string, mysqlUsername string, mysqlPassword string, csvFoldername string) error {
	if exportType == "mysql" {
		err := ConnectMySQL(mysqlServerURL, mysqlUsername, mysqlPassword)

		if err != nil {
			return err
		}

		return InitializeMySQLDatabase()
	} else if exportType == "csv" {
		return SetCSVFolder(csvFoldername)
	} else if exportType == "json" {
		return nil
	}

	return fmt.Errorf("invalid exporter type: %s", exportType)
}

func ExportAll(cache DiscourseCache, exportType string) {
	ExportUsers(cache.Users, exportType)
	ExportTopicComments(cache.Topics, exportType)
	ExportTopicEdits(cache.TopicEdits, exportType)
}

func ExportUsers(users map[string]*discourse.TopicParticipant, exportType string) {
	userEntries := userMapToUserEntry(users)

	if exportType == "mysql" {
		ExportUsersMySQL(userEntries)
	} else if exportType == "csv" {
		ExportUsersCSV(userEntries)
	} else if exportType == "json" {
		ExportUsersJSON(userEntries)
	}
}

func ExportTopicComments(topics map[string]map[int]*discourse.TopicData, exportType string) {
	topicComments := topicMapToTopicComments(topics)

	if exportType == "mysql" {
		ExportTopicCommentsMySQL(topicComments)
	} else if exportType == "csv" {
		ExportTopicCommentsCSV(topicComments)
	} else if exportType == "json" {
		ExportTopicCommentsJSON(topicComments)
	}
}

func ExportTopicEdits(revisions map[int]map[int]*discourse.PostRevision, exportType string) {
	topicEdits := topicRevisionMapToTopicEdits(revisions)

	if exportType == "mysql" {
		ExportTopicEditsMySQL(topicEdits)
	} else if exportType == "csv" {
		ExportTopicEditsCSV(topicEdits)
	} else if exportType == "json" {
		ExportTopicEditsJSON(topicEdits)
	}
}

func userMapToUserEntry(users map[string]*discourse.TopicParticipant) (userEntries []UserEntry) {
	for _, participant := range users {
		userEntries = append(userEntries, UserEntry{
			UserID:           participant.ID,
			Username:         participant.Username,
			Name:             participant.Name,
			PrimaryGroupName: participant.PrimaryGroupName,
		})
	}

	return userEntries
}

func topicMapToTopicComments(topics map[string]map[int]*discourse.TopicData) (topicComments []TopicCommentsEntry) {
	for category_slug, topic_list := range topics {
		for topic_id, topic := range topic_list {
			for postNum, post := range topic.PostStream.Posts {
				topicComments = append(topicComments, TopicCommentsEntry{
					CategorySlug:  category_slug,
					TopicID:       topic_id,
					PostID:        post.ID,
					CreationTime:  post.CreatedAt,
					UpdateTime:    post.UpdatedAt,
					Username:      post.Username,
					IsInitialPost: postNum == 0,
				})
			}
		}
	}

	return topicComments
}

func topicRevisionMapToTopicEdits(revisions map[int]map[int]*discourse.PostRevision) (topicEdits []TopicEditsEntry) {
	for topic_id, topicRevisions := range revisions {
		for revision_index, topicRevision := range topicRevisions {
			topicEdits = append(topicEdits, TopicEditsEntry{
				TopicID:      topic_id,
				EditNumber:   revision_index,
				CreationTime: topicRevision.CreatedAt,
				Username:     topicRevision.Username,
			})
		}
	}

	return topicEdits
}
