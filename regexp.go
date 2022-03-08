package main

import (
	"log"
	"regexp"
)

func regexpSelectN(n int, regexps []string, names []string) (matches []string) {
	if n <= 0 {
		matches = make([]string, 0)
	} else {
		matches = make([]string, 0, n)
	}

	res := make([]*regexp.Regexp, 0, len(regexps))
	for _, reStr := range regexps {
		re, err := regexp.Compile(reStr)
		if err != nil {
			log.Printf("warning: invalid regexp ignored: %q: %v", reStr, err)
			continue
		}
		res = append(res, re)
	}

namesLoop:
	for _, name := range names {
		if len(matches) == n {
			break
		}
		for _, re := range res {
			if re.MatchString(name) {
				matches = append(matches, name)
				continue namesLoop
			}
		}
	}

	return
}
