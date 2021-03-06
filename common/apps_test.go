package common_test

import (
	"testing"

	"github.com/datacol-io/datacol/common"
	"github.com/stretchr/testify/assert"
)

func TestMergeAppDomains(t *testing.T) {
	testcases := []struct {
		domains  []string
		item     string
		expected []string
	}{
		{[]string{"a.com"}, "a.com", []string{"a.com"}},
		{[]string{"a.com"}, "b.com", []string{"a.com", "b.com"}},
		{[]string{}, ":b.com", []string{}},
		{[]string{"a.com", "b.com"}, ":a.com", []string{"b.com"}},
	}

	for _, tc := range testcases {
		t.Run(tc.item, func(t *testing.T) {
			assert.Equal(t, common.MergeAppDomains(tc.domains, tc.item), tc.expected, "Should get correct domains")
		})
	}
}
