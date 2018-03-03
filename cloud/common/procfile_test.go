package common_test

import (
	"testing"

	"github.com/datacol-io/datacol/cloud/common"
	"github.com/stretchr/testify/assert"
)

func TestParseProcfile(t *testing.T) {
	testcases := []struct {
		name string
		in   []byte
		out  common.Procfile
	}{
		{
			"std-1",
			[]byte(`---
web: ./bin/web`),
			common.StdProcfile{
				"web": "./bin/web",
			},
		},

		{
			"extended-1",
			[]byte(`---
web:
  command: ./bin/web`),
			common.ExtProcfile{
				"web": common.Process{
					Command: "./bin/web",
				},
			},
		},
		{
			"extended-2",
			[]byte(`---
web:
  command: [npm, start]
timer:
  command: sleep 5m`),
			common.ExtProcfile{
				"timer": common.Process{
					Command: "sleep 5m",
				},
				"web": common.Process{
					Command: []interface{}{"npm", "start"},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := common.ParseProcfile(tc.in)
			assert.NoError(t, err)
			assert.Equal(t, tc.out, out)
		})
	}

}
