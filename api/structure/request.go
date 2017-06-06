package structure

// ConfigRequest is config
type ConfigRequest struct {
	Action string
}

// ZoneRequest is zone 
type ZoneRequest struct {
	Domain string
}

// ZoneDynamicGroupRequest is zone dynamic group 
type ZoneDynamicGroupRequest struct {
        DynamicGroupName string
}

// ZoneDynamicGroupDynamicRecordForceDownRequest is zone dynamic group 
type ZoneDynamicGroupDynamicRecordForceDownRequest struct {
        ForceDown bool
}
