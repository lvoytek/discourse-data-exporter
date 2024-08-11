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
	IsInitialPost bool      `csv:"Is the topic's text" json:"is_initial_post"`
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
