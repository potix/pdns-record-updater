package watcher




func (w *Watcher) icmpCheck(host string, retryCount int, retryWait uint16, timeout uint16) (alive bool) {
	ipList, err := net.LookupIP(host)
	if err != nil {
		// XX logger
		return false
	}
	for _, ip := range ipList {
		pingOk := false
		for i := 0; i < retryCount; i++ {
			ipv := 0
			var conn *icmp.PacketConn
			var err error
			switch len([]byte(ip)) {
			case 4:
				ipv = 4
				conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
			case 16:
				ipv = 6
				conn, err = icmp.ListenPacket("ip6:icmp", "::")
			default:
				// XXX logger unspported protocol version
				continue
			}
			if err != nil {
				if retryWait > 0 {
					time.Sleep(time.Duration(retryWait))
				}
				continue
			}
			echoReq := &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  int(w.icmpSeq.GetAndIncrement()),
				Data: []byte("Are you alive?"),
			}
			// ipListとretryCountがそこまで多くないこと想定
			defer conn.Close()
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
				if retryWait > 0 {
					time.Sleep(time.Duration(retryWait))
				}
				continue
			}
			if _, err := conn.WriteTo(wb, &net.IPAddr{IP: ip}); err != nil {
				if retryWait > 0 {
					time.Sleep(time.Duration(retryWait))
				}
				continue
			}
			rb := make([]byte, 1500)
			if err := conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second)); err != nil {
				if retryWait > 0 {
					time.Sleep(time.Duration(retryWait))
				}
				continue
			}
		Read:
			n, _ /* peer */, err := conn.ReadFrom(rb)
			if err != nil {
				if retryWait > 0 {
					time.Sleep(time.Duration(retryWait))
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
			rm, err := icmp.ParseMessage(proto, rb[:n])
			if err != nil {
				if retryWait > 0 {
					time.Sleep(time.Duration(retryWait))
				}
				continue
			}
			switch rm.Type {
			case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
				echoReply := (rm.Body).(*icmp.Echo)
				if echoReply.ID != echoReq.ID || echoReply.Seq != echoReq.Seq {
					goto Read
					break
				}
				pingOk = true
			default:
				goto Read
			}
		}
		if !pingOk {
			return false
		}
	}
	return true
}

