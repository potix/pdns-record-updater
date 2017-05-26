package structure

// StaticRecordResultResponse is record result
type StaticRecordResultResponse struct {
        Name    string
        Type    string
        TTL     uint32
        Content string
}

// DynamicRecordResultResponse is dynamic record result
type DynamicRecordResultResponse struct {
        Name    string
        Type    string
        TTL     uint32
        Content string
        Alive   bool
}

// ZoneResultResponse is zone result
type ZoneResultResponse struct {
	NameServer    []*StaticRecordResultResponse
	StaticRecord  []*StaticRecordResultResponse
	DynamicRecord []*DynamicRecordResultResponse
}

// WatchResultResponse is result
type WatchResultResponse struct {
	Zone map[string]*ZoneResultResponse
}

// RecordResponse is RecordResponse
type RecordResponse struct {
	NameServer    []*StaticRecordResultResponse
	StaticRecord  []*StaticRecordResultResponse
	DynamicRecord []*DynamicRecordResultResponse
}
