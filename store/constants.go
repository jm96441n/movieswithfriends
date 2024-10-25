package store

import "errors"

// errors
var (
	ErrNoRecord                        = errors.New("store: no matching record found")
	ErrMemberPartyCombinationNotUnique = errors.New("store: member party combination not unique")
	ErrDuplicatePartyName              = errors.New("party name already exists")
	ErrDuplicatePartyShortID           = errors.New("party short id already exists")
)

type WatchStatusEnum string

const (
	WatchStatusUnwatched WatchStatusEnum = "unwatched"
	WatchStatusSelected  WatchStatusEnum = "selected"
	WatchStatusWatched   WatchStatusEnum = "watched"
)

const pgUniqueViolationCode = "23505"
