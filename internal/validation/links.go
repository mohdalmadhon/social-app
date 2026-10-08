package validation

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type linkRule struct {
	label     string
	hosts     []string
	needsPath bool
}

var linkRules = map[string]linkRule{
	"website":   {label: "website", hosts: nil, needsPath: false},
	"linkedin":  {label: "LinkedIn", hosts: []string{"linkedin.com"}, needsPath: true},
	"twitter":   {label: "Twitter / X", hosts: []string{"twitter.com", "x.com"}, needsPath: true},
	"instagram": {label: "Instagram", hosts: []string{"instagram.com", "instagr.am"}, needsPath: true},
}

var (
	schemeRegex = regexp.MustCompile(`(?i)^[a-z][a-z0-9+.\-]*://`)
	hostRegex   = regexp.MustCompile(`^([a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?\.)+([a-z]{2,24}|xn--[a-z0-9\-]{2,59})$`)
)

func IsLinkField(field string) bool {
	_, ok := linkRules[field]
	return ok
}

func NormalizeLink(field string, value string) (string, error) {
	rule, ok := linkRules[field]
	if !ok {
		return "", errors.New("invalid link field")
	}

	value = strings.TrimSpace(value)

	if value == "" {
		return "", nil
	}

	if strings.ContainsAny(value, " \t\r\n") {
		return "", fmt.Errorf("%s link cannot contain spaces", rule.label)
	}

	if !schemeRegex.MatchString(value) {
		value = "https://" + value
	}

	parsed, err := url.Parse(value)

	if err != nil {
		return "", fmt.Errorf("%s link is not a valid URL", rule.label)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%s link must start with http:// or https://", rule.label)
	}

	if parsed.User != nil {
		return "", fmt.Errorf("%s link cannot contain credentials", rule.label)
	}

	host := strings.ToLower(parsed.Hostname())

	if !hostRegex.MatchString(host) {
		return "", fmt.Errorf("%s link is not a valid URL", rule.label)
	}

	if rule.hosts != nil {
		if parsed.Port() != "" {
			return "", fmt.Errorf("%s link is not a valid URL", rule.label)
		}

		matched := false

		for _, allowed := range rule.hosts {
			if host == allowed || strings.HasSuffix(host, "."+allowed) {
				matched = true
				break
			}
		}

		if !matched {
			return "", fmt.Errorf("%s link must be a %s address (%s)", rule.label, rule.label, strings.Join(rule.hosts, " or "))
		}

		if rule.needsPath && strings.Trim(parsed.Path, "/") == "" {
			return "", fmt.Errorf("%s link must point to your profile", rule.label)
		}
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""

	result := parsed.String()

	if RuneLen(result) > MaxAboutField {
		return "", fmt.Errorf("%s link cannot be more than %d characters", rule.label, MaxAboutField)
	}

	return result, nil
}
