package util

import (
	"regexp"
	"strings"
)

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
const delimiter string = "-+"

var delimiterRx = regexp.MustCompile(delimiter)
var dns1123LabelRx = regexp.MustCompile(dns1123LabelFmt)

const telepresenceMaxLength int = 57

func ToValidName(name string) string {
	name = strings.ToLower(name)
	invalidString := dns1123LabelRx.ReplaceAllString(name, "")
	invalidChar := strings.Split(invalidString, "")
	for _, i := range invalidChar {
		name = strings.Replace(name, i, "-", -1)
	}
	if len(name) > telepresenceMaxLength {
		return name[:57]
	}

	name = delimiterRx.ReplaceAllString(name, "-")
	if strings.HasSuffix(name, "-") {
		name = name[:len(name)-1]
	}
	return name
}
