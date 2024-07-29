package main

import "time"

// Metric Data
type TopicCommentsEntry struct {
	category_slug   string
	topic_id        int
	post_id         int
	creation_time   time.Time
	update_time     time.Time
	username        string
	is_initial_post bool
}

type TopicEditsEntry struct {
	topic_id      int
	edit_number   int
	creation_time time.Time
	username      string
}

// Context Data
type UserEntry struct {
	user_id            int
	username           string
	name               string
	primary_group_name string
}
