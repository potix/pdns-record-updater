package structure

// RecordResultResponse is record result
type RecordResultResponse struct {
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
        Alive   uint32
}

// ZoneResultResponse is zone result
type ZoneResultResponse struct {
	NameServer    []*RecordResult
	Record        []*RecordResult
	DynamicRecord []*DynamicRecordResult
}

// WatchResultResponse is result
type WatchResultResponse struct {
	Zone map[string]*ZoneResult
}

// RecordResponse is RecordResponse
type RecordResponse struct {
	NameServer    []*RecordResult
	Record        []*RecordResult
	DynamicRecord []*DynamicRecordResult
}
