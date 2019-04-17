package isdomain

import "regexp"

// DomainRegexpStr is a regular expression string to validate domains.
const DomainRegexpStr = "^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])(\\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9]))*$"

var domainRegexp *regexp.Regexp

func init() {
	domainRegexp = regexp.MustCompile(DomainRegexpStr)
}
