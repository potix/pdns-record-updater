package structure

import (
	"github.com/potix/pdns-record-updater/contexter"
)

// ZoneRequest is zone 
type ZoneRequest struct {
	Domain string
}

// ZoneDynamicGroupRequest is zone dynamic group 
type ZoneDynamicGroupRequest struct {
        DynamicGroupName string
}
