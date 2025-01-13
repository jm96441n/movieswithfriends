package store

import "errors"

type WatchStatusEnum string

const pgUniqueViolationCode = "23505"

// errors
var (
	ErrNoRecord                        = errors.New("store: no matching record found")
	ErrMemberPartyCombinationNotUnique = errors.New("store: member party combination not unique")
	ErrDuplicatePartyName              = errors.New("party name already exists")
	ErrDuplicatePartyShortID           = errors.New("party short id already exists")
	ErrDuplicateEmailAddress           = errors.New("email address already exists")
)

const (
	WatchStatusUnwatched WatchStatusEnum = "unwatched"
	WatchStatusSelected  WatchStatusEnum = "selected"
	WatchStatusWatched   WatchStatusEnum = "watched"
)
