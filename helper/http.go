package helper

import(
        "net"
        "net/http"
        "crypto/tls"
        "time"
)

func newHTTPTransport() (transport *http.Transport) {
        return &http.Transport{
                Proxy: http.ProxyFromEnvironment,
                Dial: (&net.Dialer{
                        Timeout:   30 * time.Second,
                        KeepAlive: 30 * time.Second,
                }).Dial,
                TLSHandshakeTimeout:   10 * time.Second,
                ExpectContinueTimeout: 1 * time.Second,
        }
}

// NewHTTPClient is new http client
func NewHTTPClient(scheme string, host string, tlsSkipVerify bool, timeout uint32) (*http.Client) {
        transport := newHTTPTransport()
        if scheme == "https" {
                transport.TLSClientConfig = &tls.Config{ServerName: host, InsecureSkipVerify: tlsSkipVerify}
        }
        return &http.Client{
                Transport: transport,
                Timeout: time.Duration(timeout) * time.Second,
        }
}
