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

// TargetRequest is config of target
type TargetRequest struct {
	TargetName     string   `json:"targetName"     yaml:"targetName"     toml:"targetName"`
        Protocol       string   `json:"protocol"       yaml:"protocol"       toml:"protocol"`       // プロトコル icmp, udp, udpRegexp, tcp, tcpRegexp, http, httpRegexp
        Dest           string   `json:"dest"           yaml:"dest"           toml:"dest"`           // 宛先
        TCPTLS         bool     `json:"tcpTls"         yaml:"tcpTls"         toml:"tcpTls"`         // TCPにTLSを使う
        HTTPMethod     string   `json:"httpMethod"     yaml:"httpMethod"     toml:"httpMethod"`     // HTTPメソッド
        HTTPStatusList []string `json:"httpStatusList" yaml:"httpStatusList" toml:"httpStatusList"` // OKとみなすHTTPステータスコード
        Regexp         string   `json:"regexp"         yaml:"regexp"         toml:"regexp"`         // OKとみなす正規表現
        ResSize        uint32   `json:"resSize"        yaml:"resSize"        toml:"resSize"`        // 受信する最大レスポンスサイズ
        Retry          uint32   `json:"retry"          yaml:"retry"          toml:"retry"`          // リトライ回数
        RetryWait      uint32   `json:"retryWait"      yaml:"retryWait"      toml:"retryWait"`      // 次のリトライまでの待ち時間
        Timeout        uint32   `json:"timeout"        yaml:"timeout"        toml:"timeout"`        // タイムアウトしたとみなす時間
        TLSSkipVerify  bool     `json:"tlsSkipVerify"  yaml:"tlsSkipVerify"  toml:"tlsSkipVerify"`  // TLSの検証をスキップする
}

// Validate is validate target request
func (t *TargetRequest) Validate() (bool) {
        if t.TargetName == "" ||  t.Protocol == "" || t.Dest == "" {
                belog.Error("no name or no protocol or no dest")
                return false
        }
        if t.Protocol == "http" || t.Protocol == "httpRegexp" {
                if t.HTTPMethod == "" || t.HTTPStatusList == nil || len(t.HTTPStatusList) == 0 {
                        belog.Error("no httpMethod or no httpStatusList")
                        return false
                }
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

