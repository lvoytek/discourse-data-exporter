package main

import (
	"log"
	"sync"

	"github.com/lvoytek/discourse_client_go/pkg/discourse"
)

type DiscourseCache struct {
	// Topics mapped by category slug and topic ID
	Topics map[string]map[int]*discourse.TopicData
	Users  map[string]*discourse.User
}

// Cache data used to avoid unnecessary Discourse API calls
var (
	cache = DiscourseCache{
		Topics: make(map[string]map[int]*discourse.TopicData),
		Users:  make(map[string]*discourse.User),
	}
	cacheWriteMutex sync.Mutex
)

func Collect(discourseClient *discourse.Client, categoryList []string) DiscourseCache {
	var collectorWg sync.WaitGroup

	for _, categorySlug := range categoryList {
		collectorWg.Add(1)
		go collectTopicsAndUsersFromCategory(&collectorWg, discourseClient, categorySlug)
	}

	collectorWg.Wait()

	return cache
}

func collectTopicsAndUsersFromCategory(wg *sync.WaitGroup, discourseClient *discourse.Client, categorySlug string) {
	defer wg.Done()
	topics, ok := cache.Topics[categorySlug]

	if !ok {
		topics = map[int]*discourse.TopicData{}
	}

	// Check each page of topics for category until there are no new topic bumps
	page := 1
	newTopics := []discourse.SuggestedTopic{}
	for {
		categoryData, err := discourse.GetCategoryContentsBySlug(discourseClient, categorySlug, page)

		if err != nil {
			log.Println("Category data collection error for", categorySlug, "on page", page, "-", err)
			return
		}

		if len(categoryData.TopicList.Topics) == 0 {
			break
		}

		// Add listed users to user map, overriding old data
		for _, newUser := range categoryData.Users {
			cache.Users[newUser.Username] = &newUser
		}

		newTopics = append(newTopics, categoryData.TopicList.Topics...)

		// Check if final topic on this page has not been updated since last check
		cachedCompareTopic, ok := topics[newTopics[len(newTopics)-1].ID]

		if ok && cachedCompareTopic.LastPostedAt.Compare(newTopics[len(newTopics)-1].LastPostedAt) >= 0 {
			break
		}

		page++
	}

	for _, topicOverview := range newTopics {
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
		cacheWriteMutex.Lock()
		defer cacheWriteMutex.Unlock()
		cache.Topics[categorySlug] = topics
	}
}
