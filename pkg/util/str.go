package util

import (
	"fmt"
	"regexp"
	"strings"
)

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
const delimiter string = "-+"

var delimiterRx = regexp.MustCompile(delimiter)
var dns1123LabelRx = regexp.MustCompile(dns1123LabelFmt)

const specialChar string = "]^\\\\[.()-"

var specialStrRx = regexp.MustCompile("[" + specialChar + "]")

const telepresenceMaxLength int = 57

func ToValidName(name string) string {
	name = strings.ToLower(name)
	invalidString := dns1123LabelRx.ReplaceAllString(name, "")
	invalidChar := strings.Split(invalidString, "")
	for _, i := range invalidChar {
		name = strings.Replace(name, i, "-", -1)
	}
	if len(name) > telepresenceMaxLength {
		name = name[:57]
	}
	name = delimiterRx.ReplaceAllString(name, "-")
	if strings.HasSuffix(name, "-") {
		name = name[:len(name)-1]
	}
	return name
}

func SpecialStr(s string) string {
	if specialStrRx.Match([]byte(s)) {
		return fmt.Sprintf("'%s'", s)
	}
	return s
}
