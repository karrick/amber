package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type configuration map[string]map[string]string

func parseConfigFile(pathname string) (conf configuration, err error) {
	fh, err := os.Open(pathname)
	if err != nil {
		return
	}
	defer fh.Close()
	buf := bufio.NewReader(fh)
	conf = make(map[string]map[string]string)
	section := "General"
	sectionRe := regexp.MustCompile("^\\[([^\\]]+)\\]$")
	keyValRe := regexp.MustCompile("^([^=]+)\\s*=\\s*(.+)$")
	for {
		var line string
		line, err = buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		line = strings.TrimSpace(line)
		switch {
		case len(line) == 0:
			continue
		case strings.IndexRune(line, ';') == 0:
			continue
		}
		if md := sectionRe.FindStringSubmatch(line); md != nil {
			section = md[1]
		} else if md := keyValRe.FindStringSubmatch(line); md != nil {
			if conf[section] == nil {
				conf[section] = make(map[string]string)
			}
			conf[section][md[1]] = md[2]
		} else {
			err = fmt.Errorf("invalid config line: %s", line)
			return
		}
	}
	return
}
