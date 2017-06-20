package structure

import (
	"github.com/potix/belog"
	"strings"
)

// ConfigRequest is config
type ConfigRequest struct {
	Action string `json:"action"`
}

// Validate is validate config request
func (c ConfigRequest) Validate() (bool) {
	if strings.ToUpper(c.Action) != "SAVE" && strings.ToUpper(c.Action) != "LOAD" {
		belog.Warn("unexpected action")
		return false
	}
	return true
}

// ZoneRequest is zone 
type ZoneRequest struct {
	PrimaryNameServer  string  `json:"primaryNameServer"`
	Email              string  `json:"email"`
	Domain             string  `json:"domain"`
}

// Validate is validate zone request
func (z ZoneRequest) Validate() (bool) {
	if z.PrimaryNameServer == "" || z.Email == "" || z.Domain == ""  {
		belog.Warn("no primaryNameServer or no email or no domain")
		return false
	}
	return true
}

// ZoneDomainRequest is zone 
type ZoneDomainRequest struct {
	PrimaryNameServer  string  `json:"primaryNameServer"`
	Email              string  `json:"email"`
}

// Validate is validate zone domain request
func (z ZoneDomainRequest) Validate() (bool) {
	if z.PrimaryNameServer == "" || z.Email == "" {
		belog.Warn("no primaryNameServer or no email")
		return false
	}
	return true
}

// ZoneDynamicGroupRequest is zone dynamic group 
type ZoneDynamicGroupRequest struct {
        DynamicGroupName string `json:"dynamicGroupName"`
}

// Validate is validate zone dynamic group request
func (z ZoneDynamicGroupRequest) Validate() (bool) {
	if z.DynamicGroupName == "" {
		belog.Warn("no dynamicGroupName")
		return false
	}
	return true
}

// ZoneDynamicGroupDynamicRecordForceDownRequest is zone dynamic group 
type ZoneDynamicGroupDynamicRecordForceDownRequest struct {
        ForceDown bool `json:"forceDown"`
}

