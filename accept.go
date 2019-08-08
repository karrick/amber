package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	defaultQualityFactor = 0
)

func parseAcceptContentType(hreq *http.Request, defaultContentType string) (contentType string) {
	// get requested content type from query and parse it
	return parseAcceptString(hreq.Header.Get("Accept"), defaultContentType)
}

func parseAcceptString(accept, defaultContentType string) (contentType string) {
	// "text/plain ; q=0.2, text/html"

	contentType = defaultContentType
	score := float64(0)

	if accept != "" {
		mediaRanges := strings.Split(accept, ",")
		for _, mediaRange := range mediaRanges {
			qualityFactor := float64(1)
			// "text/plain ; q=0.2"
			typ := strings.Split(mediaRange, ";")
			if len(typ) == 2 {
				// "q=0.2"
				// TODO: what does "level=2" mean?
				param := strings.Split(typ[1], "=")
				qv := strings.TrimSpace(param[1])
				var err error
				if qualityFactor, err = strconv.ParseFloat(qv, 32); err != nil {
					log.Printf("WARNING: cannot parse quality value: %#v", mediaRange)
					qualityFactor = defaultQualityFactor // ???
				}
			}
			if qualityFactor > score {
				contentType = strings.TrimSpace(typ[0])
				score = qualityFactor
			}
		}
	}

	return
}
