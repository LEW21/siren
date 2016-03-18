package main

import (
	"strings"
)

func SplitSubPath(path string) (string, string) {
	subpath := ""
	if strings.Contains(path, ":") {
		splitted := strings.SplitN(path, ":", 2)
		path = splitted[0]
		subpath = splitted[1]
		if len(subpath) > 0 && subpath[0] == '/' {
			subpath = subpath[1:]
		}
		if subpath == "." {
			subpath = ""
		}
	}
	return path, subpath
}
