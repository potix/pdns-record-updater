package watcher

import (
        "github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"github.com/potix/pdns-record-updater/configurator"
	"sync/atomic"
	"time"
)

var seq uint32 = 0xFFFFFFFF;

type icmpWatcher struct {
	ipAddr     string
	retry      uint32
	retryWait  uint32
	timeout    uint32
}

func (i *icmpWatcher) getSeqNumber() (uint32) {
	return atomic.AddUint32(&seq, 1);
}

func (i *icmpWatcher) sendIcmp(ip net.IP) (uint32) {
	ipv := 0
	var err error
	switch len([]byte(ip)) {
	case 4:
		ipv = 4
		conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	case 16:
		ipv = 6
		conn, err := icmp.ListenPacket("ip6:icmp", "::")
	default:
		return 0, fmt.Errorf("unsupported protocol version (%v)", i.ipAddr)
	}
	if err != nil {
		belog.Notice("can not create icmp connection (%v)", i.ipAddr)
		if i.retryWait > 0 {
			time.Sleep(time.Duration(i.retryWait))
		}
		return 0
	}
	defer conn.Close()
	echoReq := &icmp.Echo{
		ID:   os.Getpid() & 0xFFFF,
		Seq:  int(i.getSeqNumber() & 0xFFFF),
		Data: []byte("Are you alive?"),
	}
	switch ipv {
	case 4:
		icmpType := ipv4.ICMPTypeEcho
	case 6:
		icmpType := ipv6.ICMPTypeEchoRequest
	default:
		panic("not reached")
	}
	wm := icmp.Message{
		Type: icmpType,
		Code: 0,
		Body: echoReq,
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		belog.Notice("can not marshal message (%v)", wm)
		if i.retryWait > 0 {
			time.Sleep(time.Duration(i.retryWait))
		}
		return 0
	}
	if _, err := conn.WriteTo(wb, &net.IPAddr{IP: ip}); err != nil {
		belog.Notice("can not write message to connection (%v)", i.ipAddr)
		if i.retryWait > 0 {
			time.Sleep(time.Duration(i.retryWait))
		}
		return 0
	}
	if err := conn.SetReadDeadline(time.Now().Add(time.Duration(i.timeout) * time.Second)); err != nil {
		belog.Notice("can not write message to connection (%v)", i.ipAddr)
		if i.retryWait > 0 {
			time.Sleep(time.Duration(i.retryWait))
		}
		return 0
	}
	if i.resSize == 0 {
		i.resSize = 512
	}
	rb := make([]byte, i.resSize)
Read:
	rlen, _ /* peer */, err := conn.ReadFrom(rb)
	if err != nil {
		// レスポンスを読み込みめなかった
		if i.retryWait > 0 {
			time.Sleep(time.Duration(i.retryWait))
		}
		continue
	}
	var proto int
	switch ipv {
	case 4:
		proto = 1 // iana.ProtocolICMP
	case 6:
		proto = 58 // iana.ProtocolIPv6ICMP
	default:
		panic("not reached")
	}
	rm, err := icmp.ParseMessage(proto, rb[:rlen])
	if err != nil {
		// レスポンスのパースに失敗した
		if i.retryWait > 0 {
			time.Sleep(time.Duration(i.retryWait))
		}
		// 読み込みからもう一度
		goto Read
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
		echoReply := (rm.Body).(*icmp.Echo)
		if echoReply.ID != echoReq.ID || echoReply.Seq != echoReq.Seq {
			goto Read
		}
	default:
		// 何か違うタイプのICMPを拾った
		// 読み込みからもう一度
		goto Read
	}
	return 1
}

func (i *icmpWatcher) isAlive() (uint32) {
	ip := net.ParseIP(i.ipAddr)
	if ip == nil {
		belog.Error("can not parse ip address (%v)", i.ipAddr)
		return 0
	}
	for i := 0; i < i.retry; i++ {
		i.sendIcmp(ip)
	}
	// retryの最大に達した
	return 0
}

func icmpWatcherNew(target *configurator.Target) (*protoWatcherIf) {
	return &icmpWatcher {
		ipAddr:    target.Dest,
		retry:     target.Retry,
		retryWait: target.RetryWait,
		timeout:   target.Timeout,
		resSize:   target.ResSize,
	}, nil
}

