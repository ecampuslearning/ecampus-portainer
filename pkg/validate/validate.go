package validate

import (
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	minURLRuneCount = 3
	maxURLRuneCount = 2083

	ipPattern           = `(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`
	urlSchemaPattern    = `((ftp|tcp|udp|wss?|https?):\/\/)`
	urlUsernamePattern  = `(\S+(:\S*)?@)`
	urlPathPattern      = `((\/|\?|#)[^\s]*)`
	urlPortPattern      = `(:(\d{1,5}))`
	urlIPPattern        = `([1-9]\d?|1\d\d|2[01]\d|22[0-3]|24\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-5]))`
	urlSubdomainPattern = `((www\.)|([a-zA-Z0-9]+([-_\.]?[a-zA-Z0-9])*[a-zA-Z0-9]\.[a-zA-Z0-9]+))`
	urlPattern          = `^` + urlSchemaPattern + `?` + urlUsernamePattern + `?` + `((` + urlIPPattern + `|(\[` + ipPattern + `\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(` + urlSubdomainPattern + `?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?` + urlPortPattern + `?` + urlPathPattern + `?$`
)

var (
	urlRegex         = regexp.MustCompile(urlPattern)
	uuidRegex        = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	hexadecimalRegex = regexp.MustCompile(`^[0-9a-fA-F]+$`)
	whitespaceRegex  = regexp.MustCompile(`^[[:space:]]+$`)
	dnsNameRegex     = regexp.MustCompile(`^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`)
)

func IsURL(urlString string) bool {
	if urlString == "" ||
		utf8.RuneCountInString(urlString) >= maxURLRuneCount ||
		len(urlString) <= minURLRuneCount ||
		strings.HasPrefix(urlString, ".") {
		return false
	}

	strTemp := urlString
	if strings.Contains(urlString, ":") && !strings.Contains(urlString, "://") {
		// support no indicated urlscheme but with colon for port number
		// http:// is appended so url.Parse will succeed, strTemp used so it does not impact rxURL.MatchString
		strTemp = "http://" + urlString
	}

	u, err := url.Parse(strTemp)
	if err != nil {
		return false
	}

	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}

	return urlRegex.MatchString(urlString)
}

func IsUUID(uuidString string) bool {
	return uuidRegex.MatchString(uuidString)
}

func IsHexadecimal(hexString string) bool {
	return hexadecimalRegex.MatchString(hexString)
}

func HasWhitespaceOnly(s string) bool {
	return len(s) > 0 && whitespaceRegex.MatchString(s)
}

func MinStringLength(s string, len int) bool {
	return utf8.RuneCountInString(s) >= len
}

func Matches(s, pattern string) bool {
	match, err := regexp.MatchString(pattern, s)
	return err == nil && match
}

func IsNonPositive(f float64) bool {
	return f <= 0
}

func InRange(val, left, right float64) bool {
	if left > right {
		left, right = right, left
	}

	return val >= left && val <= right
}

func IsHost(s string) bool {
	return IsIP(s) || IsDNSName(s)
}

func IsIP(s string) bool {
	return net.ParseIP(s) != nil
}

func IsDNSName(s string) bool {
	if s == "" || len(strings.ReplaceAll(s, ".", "")) > 255 {
		// constraints already violated
		return false
	}

	return !IsIP(s) && dnsNameRegex.MatchString(s)
}
