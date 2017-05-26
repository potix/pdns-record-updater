package structure

// UpdateTTLRecordRequest is update record request
type UpdateTTLRecordRequest struct {
	TTL       uint32 // TTLの更新
}

// UpdateContentRecordRequest is update record request
type UpdateContentRecordRequest struct {
	Content   string // Contentの更新
}

// UpdateAliveRecordRequest is update record request
type UpdateAliveRecordRequest struct {
	Alive     bool   // Aliveの更新
}

// UpdateForceDownRecordRequest is update record request
type UpdateForceDownRecordRequest struct {
	ForceDown bool   // ForceDownの更新
}

// RecordRequest is record request
type RecordRequest struct {
	Time       uint64 // 現在時刻を入れる。一定以上ずれてるとエラーを返す
	Domain     string // ドメイン名
	RecordKind string // レコード種別 NameServer, Static, Dynamic のいづれか
	Name       string // 名前
	Type       string // タイプ
	Content    string // Content
	UpdateTTL       *UpdateTTLRecordRequest
	UpdateContent   *UpdateContentRecordRequest
	UpdateAlive     *UpdateAliveRecordRequest
	UpdateForceDown *UpdateForceDownRecordRequest
}

