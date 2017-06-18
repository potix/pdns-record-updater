package structure

// StaticRecordWatchResultResponse is static record watch result
type StaticRecordWatchResultResponse struct {
        Name    string `json:"name"`
        Type    string `json:"type"`
        TTL     int32  `json:"ttl"`
        Content string `json:"content"`
}

// NameServerRecordWatchResultResponse is name server record watch result
type NameServerRecordWatchResultResponse struct {
        Name    string `json:"name"`
        Type    string `json:"type"`
        TTL     int32  `json:"ttl"`
        Content string `json:"content"`
}

// DynamicRecordWatchResultResponse is dynamic record watch result
type DynamicRecordWatchResultResponse struct {
        Name    string `json:"name"`
        Type    string `json:"type"`
        TTL     int32  `json:"ttl"`
        Content string `json:"content"`
        Alive   bool   `json:"alive"`
}

// ZoneWatchResultResponse is zone watch result
type ZoneWatchResultResponse struct {
        PrimaryNameServer string                                 `json:"primaryNameServer"`
        Email             string                                 `json:"email"`
	NameServerList    []*NameServerRecordWatchResultResponse `json:"nameServerList"`
	StaticRecordList  []*StaticRecordWatchResultResponse     `json:"staticRecordList"`
	DynamicRecordLst  []*DynamicRecordWatchResultResponse    `json:"dynamicRecordList"`
}

// WatchResultResponse is watch result
type WatchResultResponse struct {
	ZoneMap map[string]*ZoneWatchResultResponse `json:"zoneMap"`
}

// ZoneDomainResponse is zone domain
type ZoneDomainResponse struct {
        PrimaryNameServer string `json:"primaryNameServer"`
        Email             string `json:"email"`
}
