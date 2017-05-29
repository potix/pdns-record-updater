package structure

import (
        "github.com/potix/pdns-record-updater/contexter"
)

// StaticRecordWatchResultResponse is static record watch result
type StaticRecordWatchResultResponse struct {
        Name    string
        Type    string
        TTL     uint32
        Content string
}

// DynamicRecordWatchResultResponse is dynamic record watch result
type DynamicRecordWatchResultResponse struct {
        Name    string
        Type    string
        TTL     uint32
        Content string
        Alive   bool
}

// ZoneWatchResultResponse is zone watch result
type ZoneWatchResultResponse struct {
	NameServer    []*StaticRecordWatchResultResponse
	StaticRecord  []*StaticRecordWatchResultResponse
	DynamicRecord []*DynamicRecordWatchResultResponse
}

// WatchResultResponse is watch result
type WatchResultResponse struct {
	Zone map[string]*ZoneWatchResultResponse
}
