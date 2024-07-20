package main

import (
	"log"
	"sync"
	"time"

	"github.com/lvoytek/discourse_client_go/pkg/discourse"
)

// Cache data used to avoid unnecessary Discourse API calls
var (
	// Cache of topics labelled by category slug and topic id
	topicsCache = map[string]map[int]*discourse.TopicData{}
)

func IntervalCollect(discourseClient *discourse.Client, categoryList []string, interval time.Duration) {
	for {
		Collect(discourseClient, categoryList)
		time.Sleep(interval)
	}
}

func Collect(discourseClient *discourse.Client, categoryList []string) {
	var collectorWg sync.WaitGroup

	for _, categorySlug := range categoryList {
		collectorWg.Add(1)
		go collectTopicsFromCategory(&collectorWg, discourseClient, categorySlug)
	}

	collectorWg.Wait()
}

func collectTopicsFromCategory(wg *sync.WaitGroup, discourseClient *discourse.Client, categorySlug string) {
	defer wg.Done()
	topics, ok := topicsCache[categorySlug]

	if !ok {
		topics = map[int]*discourse.TopicData{}
	}

	//TODO: Collect additional topics when there are more than 30
	categoryData, err := discourse.GetCategoryContentsBySlug(discourseClient, categorySlug)

	if err != nil {
		log.Println("Category data collection error for", categorySlug, "-", err)
		return
	}

	for _, topicOverview := range categoryData.TopicList.Topics {
		cachedTopic, topicExists := topics[topicOverview.ID]

		// If cached topic data exists, check if it actually needs to be updated
		if topicExists && cachedTopic.LastPostedAt.Compare(topicOverview.LastPostedAt) >= 0 {
			continue
		}

		// Get a new copy of the topic
		updatedTopic, err := discourse.GetTopicByID(discourseClient, topicOverview.ID)

		if err == nil {
			topics[topicOverview.ID] = updatedTopic
		} else {
			log.Println("Download topic error:", err)
		}
	}

	if len(topics) > 0 {
		topicsCache[categorySlug] = topics
	}
}
