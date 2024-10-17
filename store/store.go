package store

import "errors"

// errors
var ErrNoRecord = errors.New("store: no matching record found")

type watchStatusEnum string

const (
	WatchStatusUnwatched watchStatusEnum = "unwatched"
	WatchStatusSelected  watchStatusEnum = "selected"
	WatchStatusWatched   watchStatusEnum = "watched"
)

const pgUniqueViolationCode = "23505"
