package main

import (
	"path/filepath"
	"testing"

	"hawx.me/code/assert"
	"hawx.me/code/mopodofo/internal/pot"
)

func TestReadTemplate(t *testing.T) {
	testcases := map[string]struct {
		entries []pot.Entry
		err     error
	}{
		"simple.txt": {
			entries: []pot.Entry{
				{
					Reference: []string{"testdata/simple.txt:1"},
					MsgID:     "simple",
				},

				{
					Reference: []string{"testdata/simple.txt:2"},
					MsgID:     "plural",
					IsPlural:  true,
				},
				{
					ExtractedComments: []string{"Personalisation available: Name, Age"},
					Reference:         []string{"testdata/simple.txt:3"},
					MsgID:             "formatted",
				},
				{
					ExtractedComments: []string{"Personalisation available: Name"},
					Reference:         []string{"testdata/simple.txt:4"},
					MsgID:             "formattedPlural",
					IsPlural:          true,
				},
				{
					Reference: []string{"testdata/simple.txt:5"},
					MsgCtxt:   "action",
					MsgID:     "save",
				},
				{
					Reference: []string{"testdata/simple.txt:6"},
					MsgCtxt:   "action",
					MsgID:     "savePlural",
					IsPlural:  true,
				},
				{
					ExtractedComments: []string{"Personalisation available: File"},
					Reference:         []string{"testdata/simple.txt:7"},
					MsgCtxt:           "action",
					MsgID:             "saveFile",
				},
				{
					ExtractedComments: []string{"Personalisation available: Directory"},
					Reference:         []string{"testdata/simple.txt:8"},
					MsgCtxt:           "action",
					MsgID:             "saveFiles",
					IsPlural:          true,
				},
			},
		},
	}

	for path, tc := range testcases {
		t.Run(path, func(t *testing.T) {
			entries, err := readTemplate(filepath.Join("testdata", path))
			assert.Equal(t, entries, tc.entries)
			assert.Equal(t, err, tc.err)
		})
	}
}
