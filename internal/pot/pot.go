// Package pot encodes GNU gettext .pot files.
package pot

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"time"
)

// A File is a .pot file.
type File struct {
	Metadata Metadata
	Entries  []Entry
}

type Metadata struct {
	CreationDate time.Time
	// Generator    string
}

type Entry struct {
	ExtractedComments []string
	Reference         []string

	MsgCtxt  string
	MsgID    string
	IsPlural bool
}

func Encode(w io.Writer, file File) error {
	buf := bufio.NewWriter(w)

	fmt.Fprintf(buf, `msgid ""
msgstr ""
"POT-Creation-Date: %s\n"
"X-Generator: mopodofo\n"
`, file.Metadata.CreationDate.Format("2006-01-02T15:04:05-07:00"))

	for _, entry := range file.Entries {
		buf.WriteByte('\n')

		for _, comment := range entry.ExtractedComments {
			buf.WriteString(wrapComment(comment))
		}

		if len(entry.Reference) > 0 {
			buf.WriteString("#: ")
			for i, ref := range entry.Reference {
				if i > 0 {
					buf.WriteByte(' ')
				}
				if strings.Contains(ref, " ") {
					buf.WriteRune('\u2068')
					buf.WriteString(ref)
					buf.WriteRune('\u2069')
				} else {
					buf.WriteString(ref)
				}
			}
			buf.WriteByte('\n')
		}

		if entry.MsgCtxt != "" {
			fmt.Fprintf(buf, "msgctxt \"%s\"\n", entry.MsgCtxt)
		}

		fmt.Fprintf(buf, "msgid \"%s\"\n", entry.MsgID)
		if entry.IsPlural {
			fmt.Fprintf(buf, "msgid_plural \"%s\"\n", entry.MsgID)
			buf.WriteString("msgstr[0] \"\"\n")
		} else {
			buf.WriteString("msgstr \"\"\n")
		}
	}

	return buf.Flush()
}

type ctxtid struct {
	ctxt, id string
}

// Merge takes a raw list of entries and merges those with the same
// msgctxt+msgid. It returns the entries sorted by msgid then msgctxt.
func Merge(entries []Entry) []Entry {
	byMsgid := map[ctxtid]Entry{}
	for _, entry := range entries {
		key := ctxtid{ctxt: entry.MsgCtxt, id: entry.MsgID}
		if old, ok := byMsgid[key]; ok {
			old.ExtractedComments = append(old.ExtractedComments, entry.ExtractedComments...)
			slices.Sort(old.ExtractedComments)
			old.ExtractedComments = slices.Compact(old.ExtractedComments)

			old.Reference = append(old.Reference, entry.Reference...)
			byMsgid[key] = old
		} else {
			byMsgid[key] = entry
		}
	}

	entries = slices.Collect(maps.Values(byMsgid))
	slices.SortFunc(entries, func(a, b Entry) int {
		return strings.Compare(a.MsgID+"/"+a.MsgCtxt, b.MsgID+"/"+b.MsgCtxt)
	})

	return entries
}

func wrapComment(text string) string {
	var l, s strings.Builder
	words := strings.Split(text, " ")

	for i, word := range words {
		if len(word)+l.Len() > 77 {
			s.WriteString("#. ")
			s.WriteString(l.String())
			s.WriteString("\n")
			l.Reset()
		}

		l.WriteString(word)
		if i < len(words)-1 {
			l.WriteString(" ")
		}
	}

	if l.Len() > 0 {
		s.WriteString("#. ")
		s.WriteString(l.String())
		s.WriteString("\n")
	}

	return s.String()
}
