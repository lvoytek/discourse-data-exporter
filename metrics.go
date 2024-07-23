package main

import "time"

type TopicCommentsEntry struct {
	category_slug string
	topic_id      int
	post_id       int
	creation_time time.Time
	update_time   time.Time
	username      string
}
