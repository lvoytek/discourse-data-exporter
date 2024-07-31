package main

import (
	"log"
	"sync"
	"time"

	"github.com/lvoytek/discourse_client_go/pkg/discourse"
)

type DiscourseCache struct {
	// Topics mapped by category slug and topic ID
	Topics     map[string]map[int]*discourse.TopicData
	Users      map[int]*discourse.TopicParticipant
	TopicEdits map[int][]*discourse.PostRevision
}

// Cache data used to avoid unnecessary Discourse API calls
var (
	cache = DiscourseCache{
		Topics:     make(map[string]map[int]*discourse.TopicData),
		Users:      make(map[int]*discourse.TopicParticipant),
		TopicEdits: make(map[int][]*discourse.PostRevision),
	}
	cacheWriteMutex sync.Mutex
)

func Collect(discourseClient *discourse.Client, categoryList []string, rateLimit time.Duration) DiscourseCache {
	var collectorWg sync.WaitGroup

	for _, categorySlug := range categoryList {
		collectorWg.Add(1)
		go collectTopicsAndUsersFromCategory(&collectorWg, discourseClient, categorySlug, rateLimit)
	}

	collectorWg.Wait()

	for _, categorySlug := range categoryList {
		collectorWg.Add(1)
		go collectTopicEditsFromCacheTopicList(&collectorWg, discourseClient, categorySlug, rateLimit)
	}

	collectorWg.Wait()

	return cache
}

func collectTopicsAndUsersFromCategory(wg *sync.WaitGroup, discourseClient *discourse.Client, categorySlug string, rateLimit time.Duration) {
	defer wg.Done()
	topics, ok := cache.Topics[categorySlug]

	if !ok {
		topics = map[int]*discourse.TopicData{}
	}

	// Check each page of topics for category until there are no new topic bumps
	page := 0
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

		newTopics = append(newTopics, categoryData.TopicList.Topics...)

		// Check if final topic on this page has not been updated since last check
		cachedCompareTopic, ok := topics[newTopics[len(newTopics)-1].ID]

		if ok && cachedCompareTopic.LastPostedAt.Compare(newTopics[len(newTopics)-1].LastPostedAt) >= 0 {
			break
		}

		page++

		time.Sleep(rateLimit)
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
			for _, participant := range updatedTopic.Details.Participants {
				cache.Users[participant.ID] = &participant
			}
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

func collectTopicEditsFromCacheTopicList(wg *sync.WaitGroup, discourseClient *discourse.Client, categorySlug string, rateLimit time.Duration) {
	defer wg.Done()
	topics, ok := cache.Topics[categorySlug]

	if !ok {
		return
	}

	// Get all new edit pages for each topic
	for topicID, topic := range topics {
		collectTopicEditsFromTopic(discourseClient, topicID, topic, rateLimit)
	}
}

func collectTopicEditsFromTopic(discourseClient *discourse.Client, topicID int, topic *discourse.TopicData, rateLimit time.Duration) {
	revisions, ok := cache.TopicEdits[topicID]

	if !ok {
		revisions = []*discourse.PostRevision{}
	}

	topicPostID := topic.PostStream.Posts[0].ID

	numRevisions, err := discourse.GetNumPostRevisionsByID(discourseClient, topicPostID)

	if err != nil {
		log.Println("Number of topic edits data collection error for", topicID, err)
	}

	// Ignore existing revisions - index of revision in array is revision # - 2 since 2 is always the first revision
	for revisionNum := len(revisions) + 2; revisionNum <= numRevisions; revisionNum++ {
		nextRevision, err := discourse.GetPostRevisionByID(discourseClient, topicPostID, revisionNum)

		if err != nil {
			log.Println("Topic edits data collection error for", topicID, "revision", revisionNum, err)
		} else {
			revisions = append(revisions, nextRevision)
		}

		time.Sleep(rateLimit)
	}

	if len(revisions) > 0 {
		cacheWriteMutex.Lock()
		defer cacheWriteMutex.Unlock()
		cache.TopicEdits[topicID] = revisions
	}
}
