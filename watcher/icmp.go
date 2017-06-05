package watcher

import (
        "github.com/pkg/errors"
        "github.com/potix/belog"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"github.com/potix/pdns-record-updater/contexter"
	"sync/atomic"
	"net"
	"time"
	"fmt"
	"os"
)

var seq uint32 = 0xFFFFFFFF;

type icmpWatcher struct {
	ipAddr     string
	retry      uint32
	retryWait  uint32
	timeout    uint32
	resSize    uint32
}

func (i *icmpWatcher) getSeqNumber() (uint32) {
	return atomic.AddUint32(&seq, 1);
}

func (i *icmpWatcher) sendIcmp(ip net.IP) (bool, bool, error) {
	ipv := 0
	var conn *icmp.PacketConn
	var err error
	if (ip.To4() != nil) {
		belog.Debug("icmp ipv4 (%v)", i.ipAddr)
		ipv = 4
		conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	} else {
		belog.Debug("icmp ipv6 (%v)", i.ipAddr)
		ipv = 6
		conn, err = icmp.ListenPacket("ip6:icmp", "::")
	}
	if err != nil {
		return false, true, errors.Wrap(err, fmt.Sprintf("can not create icmp connection (%v)", i.ipAddr))
	}
	defer conn.Close()
	echoReq := &icmp.Echo{
		ID:   os.Getpid() & 0xFFFF,
		Seq:  int(i.getSeqNumber() & 0xFFFF),
		Data: []byte("Are you alive?"),
	}
	var icmpType icmp.Type
	switch ipv {
	case 4:
		icmpType = ipv4.ICMPTypeEcho
	case 6:
		icmpType = ipv6.ICMPTypeEchoRequest
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
		return false, false, errors.Wrap(err, fmt.Sprintf("can not marshal message (%v)", wm))
	}
	if _, err := conn.WriteTo(wb, &net.IPAddr{IP: ip}); err != nil {
		return false, true, errors.Wrap(err, fmt.Sprintf("can not write message (%v)", i.ipAddr))
	}
	if err := conn.SetReadDeadline(time.Now().Add(time.Duration(i.timeout) * time.Second)); err != nil {
		return false, false, errors.Wrap(err, fmt.Sprintf("can not set deadline (%v)", i.ipAddr))
	}
	if i.resSize == 0 {
		i.resSize = 512
	}
	rb := make([]byte, i.resSize)
Read:
	rlen, _ /* peer */, err := conn.ReadFrom(rb)
	if err != nil {
		return false, true, errors.Wrap(err, fmt.Sprintf("can not read response (%v)", i.ipAddr))
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
		belog.Notice("can not parse response (%v)", i.ipAddr)
		goto Read
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
		echoReply := (rm.Body).(*icmp.Echo)
		if echoReply.ID != echoReq.ID || echoReply.Seq != echoReq.Seq {
			belog.Debug("id or seq mismtach (%v <> %v) (%v <> %v)", echoReply.ID, echoReq.ID, echoReply.Seq, echoReq.Seq)
			goto Read
		}
	default:
		belog.Debug("unexpected icmp type (%v)", rm.Type)
		goto Read
	}
	belog.Debug("icmp ok (%v)", i.ipAddr)
	return true, false, nil
}

func (i *icmpWatcher) isAlive() (bool) {
	ip := net.ParseIP(i.ipAddr)
	if ip == nil {
		belog.Error("can not parse ip address (%v)", i.ipAddr)
		return false
	}
	var j uint32
	for j = 0; j <= i.retry; j++ {
                alive, retryable, err := i.sendIcmp(ip)
                if err != nil {
                        belog.Error("%v", err)
                }
                if !retryable {
                        return alive
                }
                if i.retryWait > 0 {
                        time.Sleep(time.Duration(i.retryWait) * time.Second)
                }
	}
        belog.Error("retry count is exceeded limit (%v)", i.ipAddr)
	return false
}

func icmpWatcherNew(target *contexter.Target) (protoWatcherIf, error) {
	return &icmpWatcher {
		ipAddr:    target.Dest,
		retry:     target.Retry,
		retryWait: target.RetryWait,
		timeout:   target.Timeout,
		resSize:   target.ResSize,
	}, nil
}

