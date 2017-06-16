package structure

// StaticRecordWatchResultResponse is static record watch result
type StaticRecordWatchResultResponse struct {
        Name    string
        Type    string
        TTL     int32
        Content string
}

// NameServerRecordWatchResultResponse is name server record watch result
type NameServerRecordWatchResultResponse struct {
        Name    string
        Type    string
        TTL     int32
        Content string
        Email   string
}

// DynamicRecordWatchResultResponse is dynamic record watch result
type DynamicRecordWatchResultResponse struct {
        Name    string
        Type    string
        TTL     int32
        Content string
        Alive   bool
}

// ZoneWatchResultResponse is zone watch result
type ZoneWatchResultResponse struct {
	NameServer    []*NameServerRecordWatchResultResponse
	StaticRecord  []*StaticRecordWatchResultResponse
	DynamicRecord []*DynamicRecordWatchResultResponse
}

// WatchResultResponse is watch result
type WatchResultResponse struct {
	Zone map[string]*ZoneWatchResultResponse
}
