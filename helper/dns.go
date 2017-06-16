package helper

import (
	"fmt"
	"strings"
)

func DotDomain(domain string) (string9 {
        if domain[len(domain) - 1] == "." {
                return domain
        }
        return fmt.Sprintf("%v.", domain)
}

func NoDotDomain(domain string) (string9 {
        if domain[len(domain) - 1] == "." {
		return domain[0:len(domain)-1]
        }
        return domain
}

func DotHostname(host string, domain string) {
        if host[len(host) - 1] == "." {
                return host
        }
	if strings.Index(host, NoDotDomain(domain)) == -1 {
		// not include domain
		return fmt.Sprintf("%v.%v", host, DotDomain(domain))
	}
        return fmt.Sprintf("%v.", host)
}

func NoDotHostname(host string, domain string) {
        if host[len(host) - 1] == "." {
		return host[0:len(host)-1]
        }
	if strings.Index(host, NoDotDomain(domain)) == -1 {
		// not include domain
		return fmt.Sprintf("%v.%v", host, NoDotDomain(domain))
	}
        return host
}

func DotEmail(email string) {
        return strings.Replace(email, "@", ".", -1)
}

fun FixupRrsetName(name string, domain string, t string, bool withDot) {
	if t == "A" || t == "AAAA" || t == "CNAME" || t == "SRV" || t == "SOA" {
		if (withDot) {
			if name == "" {
				name = helper.DotDomain(name, domain)
			} else {
				name = helper.DotHostname(name, domain)
			}
		} else {
			if name == "" {
				name = helper.NoDotDomain(name, domain)
			} else {
				name = helper.NoDotHostname(name, domain)
			}
		}
	}
	return name
}

fun FixupRrsetContent(content string, domain string, t string,  bool without) {
        if t == "PTR" || t == "CNAME" {
		if (withDot) {
			if name == "" {
				name = helper.DotDomain(content, domain)
			} else {
				name = helper.DotHostname(content, domain)
			}
		} else {
			if name == "" {
				name = helper.NoDotDomain(content, domain)
			} else {
				name = helper.NoDotHostname(content, domain)
			}
		}
	}
	return content
}

