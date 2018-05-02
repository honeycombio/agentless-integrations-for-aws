package main

import (
	"regexp"

	"github.com/sirupsen/logrus"
)

// filterKey returns true if the specified Key is not included in matchPatterns
// or is included in filterPatterns. Supports simple regex with
// https://golang.org/pkg/regexp/
func filterKey(key string, matchPatterns, filterPatterns []string) bool {
	filter := true
	var err error
	for _, pattern := range matchPatterns {
		filter, err = regexp.MatchString(pattern, key)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"key":     key,
				"pattern": pattern,
			}).Warn("unable to check key against pattern, skipping")
			continue
		}
		// If we match here, we want to include this file
		filter = !filter
		if !filter {
			break
		}
	}

	// only check filterPatterns if we matched above
	if !filter {
		for _, pattern := range filterPatterns {
			filter, err = regexp.MatchString(pattern, key)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"key":     key,
					"pattern": pattern,
				}).Warn("unable to check key against pattern, skipping")
				continue
			}
			if filter {
				break
			}
		}
	}

	return filter
}
