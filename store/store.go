package store

import "errors"

// errors
var ErrNoRecord = errors.New("store: no matching record found")

type WatchStatusEnum string

const (
	WatchStatusUnwatched WatchStatusEnum = "unwatched"
	WatchStatusSelected  WatchStatusEnum = "selected"
	WatchStatusWatched   WatchStatusEnum = "watched"
)

const pgUniqueViolationCode = "23505"
