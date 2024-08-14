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
	Users      map[string]*discourse.TopicParticipant
	TopicEdits map[int]map[int]*discourse.PostRevision
}

// Cache data used to avoid unnecessary Discourse API calls
var (
	cache = DiscourseCache{
		Topics:     make(map[string]map[int]*discourse.TopicData),
		Users:      make(map[string]*discourse.TopicParticipant),
		TopicEdits: make(map[int]map[int]*discourse.PostRevision),
	}
	cacheWriteMutex sync.Mutex
)

func Collect(discourseClient *discourse.Client, itemsToExport ItemsToExport, rateLimit time.Duration) DiscourseCache {
	var collectorWg sync.WaitGroup

	categoryList := []string{itemsToExport.LimitToCategorySlug}

	if itemsToExport.TopicComments || itemsToExport.TopicEdits {
		for _, categorySlug := range categoryList {
			collectorWg.Add(1)
			go collectTopicsAndUsersFromCategory(&collectorWg, discourseClient, categorySlug, rateLimit)
		}

		collectorWg.Wait()
	}

	if itemsToExport.TopicEdits {
		for _, categorySlug := range categoryList {
			collectorWg.Add(1)
			go collectTopicEditsFromCacheTopicList(&collectorWg, discourseClient, categorySlug, rateLimit)
		}

		collectorWg.Wait()
	}

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

	additionalUsers := map[string]*discourse.TopicParticipant{}

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
				additionalUsers[participant.Username] = &participant
			}

			// Fail safe if post creators are not in participant list
			for _, post := range updatedTopic.PostStream.Posts {
				_, userExistsInCache := cache.Users[post.Username]
				_, userExistsInAdditional := additionalUsers[post.Username]

				if !userExistsInCache && !userExistsInAdditional {

					newUser, err := discourse.GetUserByUsername(discourseClient, post.Username)

					if err != nil {
						log.Println("Could not find post creator by username ", post.Username, "-", err)
						continue
					}

					additionalUsers[newUser.User.Username] = &discourse.TopicParticipant{
						ID:               newUser.User.ID,
						Username:         newUser.User.Username,
						Name:             newUser.User.Name,
						PrimaryGroupName: newUser.User.PrimaryGroupName,
					}
				}
			}

		} else {
			log.Println("Download topic error:", err)
		}
	}

	if len(topics) > 0 {
		cacheWriteMutex.Lock()
		defer cacheWriteMutex.Unlock()
		cache.Topics[categorySlug] = topics

		// Add newly found users
		for username, additionalUser := range additionalUsers {
			_, userExists := cache.Users[username]

			if !userExists {
				cache.Users[username] = additionalUser
			}
		}

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
		revisions = map[int]*discourse.PostRevision{}
	}

	topicPostID := topic.PostStream.Posts[0].ID

	numRevisions, err := discourse.GetNumPostRevisionsByID(discourseClient, topicPostID)

	if err != nil {
		log.Println("Number of topic edits data collection error for", topicID, err)
	}

	// Update revisions by traversing through NextRevision linked list
	currentRevisionNum := 2
	for revisionCount := 1; revisionCount < numRevisions; revisionCount++ {
		nextRevision, err := discourse.GetPostRevisionByID(discourseClient, topicPostID, currentRevisionNum)

		if err != nil {
			log.Println("Topic edits data collection error for", topicID, "revision", currentRevisionNum, err)
			break
		}

		revisions[currentRevisionNum] = nextRevision
		currentRevisionNum = nextRevision.NextRevision

		time.Sleep(rateLimit)
	}

	if len(revisions) > 0 {
		cacheWriteMutex.Lock()
		defer cacheWriteMutex.Unlock()
		cache.TopicEdits[topicID] = revisions
	}
}
