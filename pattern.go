package router

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/olegshs/router/helpers"
)

type pattern string

var (
	paramRegexp = regexp.MustCompile(`{([A-Za-z_][0-9A-Za-z_]*)(\.\.\.)?}`)
)

func (p pattern) paramNames() helpers.Slice[string] {
	a := p.paramNamesMatch()
	names := make([]string, len(a))

	for i, m := range a {
		names[i] = m[1]
	}

	return names
}

func (p pattern) paramNamesMatch() [][]string {
	return paramRegexp.FindAllStringSubmatch(string(p), -1)
}

func (p pattern) httpRouterString() string {
	s := string(p)
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "*", "")

	a := paramRegexp.FindAllStringSubmatch(s, -1)

	for i, m := range a {
		var repl string
		if m[2] == "..." {
			repl = fmt.Sprintf("*%d", i)
		} else {
			repl = fmt.Sprintf(":%d", i)
		}

		s = strings.ReplaceAll(s, m[0], repl)
	}

	return s
}
