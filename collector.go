package main

import (
	"log"
	"sync"

	"github.com/lvoytek/discourse_client_go/pkg/discourse"
)

type DiscourseCache struct {
	// Topics mapped by category slug and topic ID
	Topics map[string]map[int]*discourse.TopicData
}

// Cache data used to avoid unnecessary Discourse API calls
var (
	cache DiscourseCache
)

func Collect(discourseClient *discourse.Client, categoryList []string) DiscourseCache {
	var collectorWg sync.WaitGroup

	for _, categorySlug := range categoryList {
		collectorWg.Add(1)
		go collectTopicsFromCategory(&collectorWg, discourseClient, categorySlug)
	}

	collectorWg.Wait()

	return cache
}

func collectTopicsFromCategory(wg *sync.WaitGroup, discourseClient *discourse.Client, categorySlug string) {
	defer wg.Done()
	topics, ok := cache.Topics[categorySlug]

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
		cache.Topics[categorySlug] = topics
	}
}
