package structure

// RecordRequest is record request
type RecordRequest struct {
	Time       uint64
	Domain     string
	RecordKind string
	Name       string
	Type       string
	Content    string
	TTL        uint32
	ForceDown  uint32
}
