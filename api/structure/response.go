package structure

import (
	"fmt"
)

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

// NameServerListWatchResultResponse is name server list
type NameServerListWatchResultResponse  []*NameServerRecordWatchResultResponse

func (n NameServerListWatchResultResponse) Len() int {
        return len(n)
}

func (n NameServerListWatchResultResponse) Swap(i, j int) {
        n[i], n[j] = n[j], n[i]
}

func (n NameServerListWatchResultResponse) Less(i, j int) bool {
    return fmt.Sprintf("%v %v", n[i].Name, n[i].Type) < fmt.Sprintf("%v %v", n[j].Name, n[j].Type)
}

// StaticRecordListWatchResultResponse is static record list
type StaticRecordListWatchResultResponse  []*StaticRecordWatchResultResponse

func (s StaticRecordListWatchResultResponse) Len() int {
        return len(s)
}

func (s StaticRecordListWatchResultResponse) Swap(i, j int) {
        s[i], s[j] = s[j], s[i]
}

func (s StaticRecordListWatchResultResponse) Less(i, j int) bool {
    return fmt.Sprintf("%v %v", s[i].Name, s[i].Type) < fmt.Sprintf("%v %v", s[j].Name, s[j].Type)
}

// DynamicRecordListWatchResultResponse is dynamic record list
type DynamicRecordListWatchResultResponse []*DynamicRecordWatchResultResponse

func (d DynamicRecordListWatchResultResponse) Len() int {
        return len(d)
}

func (d DynamicRecordListWatchResultResponse) Swap(i, j int) {
        d[i], d[j] = d[j], d[i]
}

func (d DynamicRecordListWatchResultResponse) Less(i, j int) bool {
    return fmt.Sprintf("%v %v", d[i].Name, d[i].Type) < fmt.Sprintf("%v %v", d[j].Name, d[j].Type)
}

// ZoneWatchResultResponse is zone watch result
type ZoneWatchResultResponse struct {
        PrimaryNameServer string                               `json:"primaryNameServer"`
        Email             string                               `json:"email"`
	NameServerList    NameServerListWatchResultResponse    `json:"nameServerList"`
	StaticRecordList  StaticRecordListWatchResultResponse  `json:"staticRecordList"`
	DynamicRecordList DynamicRecordListWatchResultResponse `json:"dynamicRecordList"`
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
