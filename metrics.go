package main

import "time"

// Metric Data
type TopicCommentsEntry struct {
	category_slug   string    `csv:"Category Slug"`
	topic_id        int       `csv:"Topic ID"`
	post_id         int       `csv:"Post ID"`
	creation_time   time.Time `csv:"Creation Time"`
	update_time     time.Time `csv:"Last Update Time"`
	username        string    `csv:"Creator Username"`
	is_initial_post bool      `csv:"Is the topic's text"`
}

type TopicEditsEntry struct {
	topic_id      int       `csv:"Topic ID"`
	edit_number   int       `csv:"Edit Number"`
	creation_time time.Time `csv:"Creation Time"`
	username      string    `csv:"Editor Username"`
}

// Context Data
type UserEntry struct {
	user_id            int    `csv:"User ID"`
	username           string `csv:"Username"`
	name               string `csv:"Name"`
	primary_group_name string `csv:"Primary Group Name"`
}
