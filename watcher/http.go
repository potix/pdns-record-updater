package watcher

func (w *Watcher) httpCheck(url string, codeList []string, bodyList []string, retryCount int, retryWait uint16, timeout uint16) (alive bool) {
	for i := 0; i < retryCount; i++ {
		httpClient := &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
		resp, err := httpClient.Get(url)
		if err != nil {
			if retryWait > 0 {
				time.Sleep(time.Duration(retryWait))
			}
			continue
		}
		if len(codeList) > 0 {
			codeMatch := false
			for _, code := range codeList {
				if regexp.GetRegexpManager().IsMatch(code, resp.Status) {
					codeMatch = true
					break
				}
			}
			if !codeMatch {
				if retryWait > 0 {
					time.Sleep(time.Duration(retryWait))
				}
				continue
			}
		}
		// retryCountがそこまで多くないこと想定
		defer resp.Body.Close()
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			if retryWait > 0 {
				time.Sleep(time.Duration(retryWait))
			}
			continue
		}
		if len(bodyList) > 0 {
			bodyMatch := false
			for _, body := range bodyList {
				if regexp.GetRegexpManager().IsMatch(body, string(responseBody)) {
					bodyMatch = true
					break
				}
			}
			if !bodyMatch {
				if retryWait > 0 {
					time.Sleep(time.Duration(retryWait))
				}
				continue
			}
		}
		return true
	}
	return false
}
