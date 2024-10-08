package main

import "time"

// Metric Data
type TopicCommentsEntry struct {
	CategorySlug  string    `csv:"Category Slug" json:"category_slug"`
	TopicID       int       `csv:"Topic ID" json:"topic_id"`
	PostID        int       `csv:"Post ID" json:"post_id"`
	CreationTime  time.Time `csv:"Creation Time" json:"creation_time"`
	UpdateTime    time.Time `csv:"Last Update Time" json:"update_time,omitempty"`
	Username      string    `csv:"Creator Username" json:"username"`
	IsInitialPost bool      `csv:"Is the topic's main post" json:"is_initial_post"`
}

type TopicEditsEntry struct {
	TopicID      int       `csv:"Topic ID" json:"topic_id"`
	EditNumber   int       `csv:"Edit Number" json:"edit_number"`
	CreationTime time.Time `csv:"Creation Time" json:"creation_time"`
	Username     string    `csv:"Editor Username" json:"username"`
}

// Context Data
type UserEntry struct {
	UserID           int    `csv:"User ID" json:"user_id"`
	Username         string `csv:"Username" json:"username"`
	Name             string `csv:"Name" json:"name,omitempty"`
	PrimaryGroupName string `csv:"Primary Group Name" json:"primary_group_name,omitempty"`
}

// All output data
type DataToExport struct {
	Posts []TopicCommentsEntry `json:"posts,omitempty"`
	Edits []TopicEditsEntry    `json:"edits,omitempty"`
	Users []UserEntry          `json:"users,omitempty"`
}

// Struct containing info on what types to export
type ItemsToExport struct {
	TopicComments bool
	TopicEdits    bool
	Users         bool

	LimitToCategorySlug string
	LimitToTopicID      int
}
