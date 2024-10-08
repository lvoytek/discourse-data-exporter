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
	cacheWriteMutex   sync.Mutex
	rateLimitMutex    sync.Mutex
	rateLimitDuration time.Duration = time.Second
)

func Collect(discourseClient *discourse.Client, itemsToExport ItemsToExport, rateLimit time.Duration) DiscourseCache {
	var collectorWg sync.WaitGroup
	rateLimitDuration = rateLimit

	categoryList := []string{itemsToExport.LimitToCategorySlug}

	// Get all categories if no category or topic specified
	if itemsToExport.LimitToCategorySlug == "" && itemsToExport.LimitToTopicID == 0 {
		allCategories, err := discourse.ListCategories(discourseClient, true)
		rateLimitDelay()

		if err != nil {
			log.Fatalln("Unable to list categories -", err)
		}

		categoryList = []string{}

		for _, nextCategory := range allCategories.CategoryList.Categories {
			categoryList = append(categoryList, nextCategory.Slug)

			for _, nextSubcategory := range nextCategory.SubcategoryList {
				categoryList = append(categoryList, nextCategory.Slug+"/"+nextSubcategory.Slug)
			}
		}
	}

	// Topic Comments and Topic Users
	if itemsToExport.TopicComments || itemsToExport.TopicEdits {
		if itemsToExport.LimitToTopicID > 0 {
			collectTopicAndAssociatedUsers(discourseClient, itemsToExport.LimitToTopicID)
		} else {
			for _, categorySlug := range categoryList {
				collectorWg.Add(1)
				go collectTopicsAndUsersFromCategory(&collectorWg, discourseClient, categorySlug)
			}

			collectorWg.Wait()
		}
	}

	// Topic Edits
	if itemsToExport.TopicEdits {
		if itemsToExport.LimitToTopicID > 0 {
			// Find single topic to export in cache
			var topicData *discourse.TopicData
			var ok bool

			for _, topics := range cache.Topics {
				topicData, ok = topics[itemsToExport.LimitToTopicID]

				if ok {
					break
				}
			}

			if ok {
				collectTopicEditsFromTopic(discourseClient, itemsToExport.LimitToTopicID, topicData)
			} else {
				log.Println("Unable to find topic", itemsToExport.LimitToTopicID, "in cache")
			}
		} else {
			for _, categorySlug := range categoryList {
				collectorWg.Add(1)
				go collectTopicEditsFromCacheTopicList(&collectorWg, discourseClient, categorySlug)
			}

			collectorWg.Wait()
		}
	}

	return cache
}

func collectTopicsAndUsersFromCategory(wg *sync.WaitGroup, discourseClient *discourse.Client, categorySlug string) {
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
		rateLimitDelay()

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
		rateLimitDelay()

		if err == nil {
			topics[topicOverview.ID] = updatedTopic

			additionalTopicUsers := getUsersListedInTopic(discourseClient, updatedTopic)

			for k, v := range additionalTopicUsers {
				additionalUsers[k] = v
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

func collectTopicAndAssociatedUsers(discourseClient *discourse.Client, topicID int) {
	updatedTopic, err := discourse.GetTopicByID(discourseClient, topicID)
	rateLimitDelay()

	if err == nil {
		additionalUsers := getUsersListedInTopic(discourseClient, updatedTopic)

		categoryData, err := discourse.ShowCategory(discourseClient, updatedTopic.CategoryID)
		rateLimitDelay()

		categoryName := ""

		if err != nil {
			log.Println("Could not find category for topic ", updatedTopic.Title, "-", err)
		} else {
			categoryName = categoryData.Category.Slug
		}

		cacheWriteMutex.Lock()
		defer cacheWriteMutex.Unlock()
		cache.Topics[categoryName] = map[int]*discourse.TopicData{topicID: updatedTopic}

		// Add newly found users
		for username, additionalUser := range additionalUsers {
			_, userExists := cache.Users[username]

			if !userExists {
				cache.Users[username] = additionalUser
			}
		}
	} else {
		log.Println("Download topic error:", err)
	}
}

func getUsersListedInTopic(discourseClient *discourse.Client, topicData *discourse.TopicData) map[string]*discourse.TopicParticipant {
	additionalUsers := map[string]*discourse.TopicParticipant{}

	for _, participant := range topicData.Details.Participants {
		additionalUsers[participant.Username] = &participant
	}

	// Fail safe if post creators are not in participant list
	for _, post := range topicData.PostStream.Posts {
		_, userExistsInCache := cache.Users[post.Username]
		_, userExistsInAdditional := additionalUsers[post.Username]

		if !userExistsInCache && !userExistsInAdditional {

			newUser, err := discourse.GetUserByUsername(discourseClient, post.Username)
			rateLimitDelay()

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

	return additionalUsers
}

func collectTopicEditsFromCacheTopicList(wg *sync.WaitGroup, discourseClient *discourse.Client, categorySlug string) {
	defer wg.Done()
	topics, ok := cache.Topics[categorySlug]

	if !ok {
		return
	}

	// Get all new edit pages for each topic
	for topicID, topic := range topics {
		collectTopicEditsFromTopic(discourseClient, topicID, topic)
	}
}

func collectTopicEditsFromTopic(discourseClient *discourse.Client, topicID int, topic *discourse.TopicData) {
	revisions, ok := cache.TopicEdits[topicID]

	if !ok {
		revisions = map[int]*discourse.PostRevision{}
	}

	topicPostID := topic.PostStream.Posts[0].ID

	numRevisions, err := discourse.GetNumPostRevisionsByID(discourseClient, topicPostID)
	rateLimitDelay()

	if err != nil {
		log.Println("Number of topic edits data collection error for", topicID, err)
	}

	if numRevisions > 1 {

		// Update revisions by traversing through linked list from latest to first
		nextRevision, err := discourse.GetPostLatestRevisionByID(discourseClient, topicPostID)
		rateLimitDelay()

		if err != nil {
			log.Println("Topic edits data collection error for", topicID, "revision latest", err)
		} else {
			currentRevisionNum := nextRevision.CurrentRevision
			for {
				revisions[currentRevisionNum] = nextRevision

				if currentRevisionNum == nextRevision.FirstRevision {
					break
				}

				currentRevisionNum = nextRevision.PreviousRevision

				nextRevision, err = discourse.GetPostRevisionByID(discourseClient, topicPostID, currentRevisionNum)
				rateLimitDelay()

				if err != nil {
					log.Println("Topic edits data collection error for", topicID, "revision", currentRevisionNum, err)
					break
				}
			}
		}
	}

	if len(revisions) > 0 {
		cacheWriteMutex.Lock()
		defer cacheWriteMutex.Unlock()
		cache.TopicEdits[topicID] = revisions
	}
}

func rateLimitDelay() {
	rateLimitMutex.Lock()
	defer rateLimitMutex.Unlock()
	time.Sleep(rateLimitDuration)
}
