package helper

import (
	"fmt"
	"strings"
)

// DotDomain is Dotted domain
func DotDomain(domain string) (string) {
        if domain[len(domain) - 1] == '.' {
                return domain
        }
        return fmt.Sprintf("%v.", domain)
}

// NoDotDomain is No dotted domain
func NoDotDomain(domain string) (string) {
        if domain[len(domain) - 1] == '.' {
		return domain[0:len(domain)-1]
        }
        return domain
}

// DotHostname is No dotted hostname
func DotHostname(host string, domain string) (string) {
        if host[len(host) - 1] == '.' {
                return host
        }
	if strings.Index(host, NoDotDomain(domain)) == -1 {
		// not include domain
		return fmt.Sprintf("%v.%v", host, DotDomain(domain))
	}
        return fmt.Sprintf("%v.", host)
}

// NoDotHostname is No dotted hostname
func NoDotHostname(host string, domain string) (string) {
        if host[len(host) - 1] == '.' {
		return host[0:len(host)-1]
        }
	if strings.Index(host, NoDotDomain(domain)) == -1 {
		// not include domain
		return fmt.Sprintf("%v.%v", host, NoDotDomain(domain))
	}
        return host
}

// DotEmail is Dotted email
func DotEmail(email string) (string) {
        return strings.Replace(email, "@", ".", -1)
}

// FixupRrsetName is fixup name of rrset
func FixupRrsetName(name string, domain string, t string, withDot bool) (string) {
	t = strings.ToUpper(t)
	if t == "A" || t == "AAAA" || t == "CNAME" || t == "SRV" || t == "SOA" {
		if (withDot) {
			if name == "" {
				name = DotDomain(domain)
			} else {
				name = DotHostname(name, domain)
			}
		} else {
			if name == "" {
				name = NoDotDomain(domain)
			} else {
				name = NoDotHostname(name, domain)
			}
		}
	}
	return name
}

// FixupRrsetContent is fixup contentof rrset
func FixupRrsetContent(content string, domain string, t string,  withDot bool) (string) {
	t = strings.ToUpper(t)
        if t == "PTR" || t == "CNAME" || t == "NS" {
		if (withDot) {
			if content == "" {
				content = DotDomain(domain)
			} else {
				content = DotHostname(content, domain)
			}
		} else {
			if content == "" {
				content = NoDotDomain(domain)
			} else {
				content = NoDotHostname(content, domain)
			}
		}
	}
	return content
}

