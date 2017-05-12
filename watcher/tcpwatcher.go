package watcher


func (w *Watcher) tcpCheck(hostPort string, retryCount int, retryWait uint16, timeout uint16) (alive bool) {
	for i := 0; i < retryCount; i++ {
		dialer := &net.Dialer{
			Timeout:   time.Duration(timeout) * time.Second,
			DualStack: true,
			Deadline:  time.Now().Add(time.Duration(timeout) * time.Second),
		}
		conn, err := dialer.Dial("tcp", hostPort)
		if err != nil {
			if retryWait > 0 {
				time.Sleep(time.Duration(retryWait))
			}
			continue
		}
		if err := conn.Close(); err != nil {
			// statistics
		}
		return true
	}
	return false
}


