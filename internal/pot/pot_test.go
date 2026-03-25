package pot

import (
	"bytes"
	"testing"
	"time"

	"hawx.me/code/assert"
)

func TestUnmarshal(t *testing.T) {
	expected := `msgid ""
msgstr ""
"POT-Creation-Date: 2025-01-02T12:13:14+00:00\n"
"X-Generator: mopodofo\n"

msgid "a"
msgstr ""

#. This is extracted
#: file/page:12
msgctxt "b-ctxt"
msgid "b"
msgstr ""

#. Usually, programs are written and documented in English, and use English at 
#. execution time for interacting with users. This is true not only from within 
#. GNU, but also in a great deal of proprietary and free software. Using a 
#. common language is quite handy for communication between developers, 
#. maintainers and users from all countries. On the other hand, most people are 
#. less comfortable with English than with their own native language, and would 
#. rather be using their mother tongue for day to day's work, as far as 
#. possible. Many would simply love seeing their computer screen showing a lot 
#. less of English, and far more of their own language.
msgid "c"
msgstr ""

#: file/a:1 file/b:2 ⁨file/c c c:4⁩
msgid "manyMessage"
msgid_plural "manyMessage"
msgstr[0] ""
`

	file := File{
		Metadata: Metadata{
			CreationDate: time.Date(2025, time.January, 2, 12, 13, 14, 0, time.UTC),
		},
		Entries: []Entry{
			{MsgID: "a"},
			{
				ExtractedComments: []string{"This is extracted"},
				Reference:         []string{"file/page:12"},
				MsgCtxt:           "b-ctxt",
				MsgID:             "b",
			},
			{
				ExtractedComments: []string{"Usually, programs are written and documented in English, and use English at execution time for interacting with users. This is true not only from within GNU, but also in a great deal of proprietary and free software. Using a common language is quite handy for communication between developers, maintainers and users from all countries. On the other hand, most people are less comfortable with English than with their own native language, and would rather be using their mother tongue for day to day's work, as far as possible. Many would simply love seeing their computer screen showing a lot less of English, and far more of their own language."},
				MsgID:             "c",
			},
			{
				Reference: []string{"file/a:1", "file/b:2", "file/c c c:4"},
				MsgID:     "manyMessage",
				IsPlural:  true,
			},
		},
	}

	var encoded bytes.Buffer
	err := Encode(&encoded, file)
	assert.Nil(t, err)
	assert.Equal(t, expected, encoded.String())
}

func TestWrapComment(t *testing.T) {
	input := "This is a paragraph wrapped to 80 characters. This is a paragraph wrapped to 80 characters. This is a paragraph wrapped to 80 characters. This is a paragraph wrapped to 80 characters."
	expected := `#. This is a paragraph wrapped to 80 characters. This is a paragraph wrapped to 
#. 80 characters. This is a paragraph wrapped to 80 characters. This is a 
#. paragraph wrapped to 80 characters.
`
	assert.Equal(t, expected, wrapComment(input))
}
