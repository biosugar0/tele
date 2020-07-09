package util

import (
	"regexp"
	"strings"
)

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

var dns1123LabelRegexp = regexp.MustCompile(dns1123LabelFmt)

const telepresenceMaxLength int = 57

func ToValidName(name string) string {
	invalidString := dns1123LabelRegexp.ReplaceAllString(name, "")
	invalidChar := strings.Split(invalidString, "")
	for _, i := range invalidChar {
		name = strings.Replace(name, i, "-", -1)
	}
	if len(name) > telepresenceMaxLength {
		return name[:57]
	}

	name = strings.Replace(name, "--", "-", -1)
	return name
}
