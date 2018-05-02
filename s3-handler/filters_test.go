package main

import (
	"testing"
)

type filterTestCase struct {
	matchPatterns  []string
	filterPatterns []string
	key            string
	shouldFilter   bool
}

func TestFilterKey(t *testing.T) {
	testCases := []filterTestCase{
		// match everything, filter nothing
		filterTestCase{
			matchPatterns:  []string{".*"},
			filterPatterns: []string{},
			key:            "some/random/key",
			shouldFilter:   false,
		},
		// filter specific key
		filterTestCase{
			matchPatterns:  []string{".*"},
			filterPatterns: []string{"some/random/key"},
			key:            "some/random/key",
			shouldFilter:   true,
		},
		// more tightly defined filter of specific key that doesn't match
		filterTestCase{
			matchPatterns:  []string{".*"},
			filterPatterns: []string{"^some/random/key$"},
			key:            "some/random/keys",
			shouldFilter:   false,
		},
		// specific match patterns that don't match our key
		filterTestCase{
			matchPatterns:  []string{"^/path1/key1.+", "^/path1/key2.+"},
			filterPatterns: []string{},
			key:            "some/random/key",
			shouldFilter:   true,
		},
		// test multiple filter patterns
		filterTestCase{
			matchPatterns:  []string{".+"},
			filterPatterns: []string{"abc", "xyz", "^some/random/key$"},
			key:            "some/random/key",
			shouldFilter:   true,
		},
		// match and exclude everything!
		filterTestCase{
			matchPatterns:  []string{".+"},
			filterPatterns: []string{".+"},
			key:            "some/random/key",
			shouldFilter:   true,
		},
	}

	for _, c := range testCases {
		filter := filterKey(c.key, c.matchPatterns, c.filterPatterns)
		if filter != c.shouldFilter {
			t.Errorf("expected %v, got %v. key: %v, matchPatterns: %v, filterPatterns: %v",
				c.shouldFilter, filter, c.key, c.matchPatterns, c.filterPatterns)
		}
	}
}
