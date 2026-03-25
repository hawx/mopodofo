// Package mo decodes gettext .mo files.
package mo

import (
	"encoding/binary"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Header struct {
	FileFormatRevision                  uint32
	NumberOfStrings                     uint32
	OffsetOfTableWithOriginalStrings    uint32
	OffsetOfTableWithTranslationStrings uint32
	SizeOfHashingTable                  uint32
	OffsetOfHashingTable                uint32
}

type LengthAndOffset struct {
	Length uint32
	Offset uint32
}

type Metadata struct {
	Language string
	NPlurals int
	Plurals  func(int) int
}

// A File stores the strings contained within a .mo file.
type File struct {
	Metadata       Metadata
	Singles        map[string]string
	Plurals        map[string][]string
	ContextSingles map[string]map[string]string
	ContextPlurals map[string]map[string][]string
}

func Decode(r io.Reader) (*File, error) {
	var cursor uint32 = 0
	var magic uint32
	err := binary.Read(r, binary.BigEndian, &magic)
	cursor += 4

	// TODO: check that I'm doing what is expected here
	var endian binary.ByteOrder
	switch magic {
	case 0xde120495:
		endian = binary.LittleEndian
	case 0x950412de:
		endian = binary.BigEndian
	default:
		return nil, fmt.Errorf("not a .mo file")
	}

	var header Header
	err = binary.Read(r, endian, &header)
	cursor += 24

	// skip to O
	skipToO := make([]byte, header.OffsetOfTableWithOriginalStrings-cursor)
	binary.Read(r, endian, &skipToO)
	cursor = header.OffsetOfTableWithOriginalStrings

	originalEntries := make([]LengthAndOffset, header.NumberOfStrings)
	for n := range header.NumberOfStrings {
		var entry LengthAndOffset
		binary.Read(r, endian, &entry)
		cursor += 8

		originalEntries[n] = entry
	}

	// skip to T
	skipToT := make([]byte, header.OffsetOfTableWithTranslationStrings-cursor)
	binary.Read(r, endian, &skipToT)
	cursor = header.OffsetOfTableWithTranslationStrings

	translationEntries := make([]LengthAndOffset, header.NumberOfStrings)
	for n := range header.NumberOfStrings {
		var entry LengthAndOffset
		binary.Read(r, endian, &entry)
		cursor += 8

		translationEntries[n] = entry
	}

	originalKeys := make([]string, header.NumberOfStrings)
	for i, entry := range originalEntries {
		skipToEntry := make([]byte, entry.Offset-cursor)
		binary.Read(r, endian, &skipToEntry)
		cursor = entry.Offset

		key := make([]byte, entry.Length)
		binary.Read(r, endian, &key)
		cursor += entry.Length

		originalKeys[i] = string(key)
	}

	translationKeys := make([]string, header.NumberOfStrings)
	for i, entry := range translationEntries {
		skipToEntry := make([]byte, entry.Offset-cursor)
		binary.Read(r, endian, &skipToEntry)
		cursor = entry.Offset

		key := make([]byte, entry.Length)
		binary.Read(r, endian, &key)
		cursor += entry.Length

		translationKeys[i] = string(key)
	}

	singles := map[string]string{}
	ctxtSingles := map[string]map[string]string{}
	plurals := map[string][]string{}
	for i := range header.NumberOfStrings {
		if strings.ContainsRune(originalKeys[i], 0x00) {
			key, _, _ := strings.Cut(originalKeys[i], "\x00")
			plurals[key] = strings.Split(translationKeys[i], "\x00")
		} else if strings.ContainsRune(originalKeys[i], 0x04) {
			ctxt, key, _ := strings.Cut(originalKeys[i], "\x04")
			if _, ok := ctxtSingles[ctxt]; !ok {
				ctxtSingles[ctxt] = map[string]string{}
			}
			ctxtSingles[ctxt][key] = translationKeys[i]
		} else {
			singles[originalKeys[i]] = translationKeys[i]
		}
	}

	metadata, err := parseMetadata(singles[""])
	if err != nil {
		return nil, err
	}
	delete(singles, "")

	return &File{
		Metadata:       metadata,
		Singles:        singles,
		Plurals:        plurals,
		ContextSingles: ctxtSingles,
	}, err
}

var pluralsRe = regexp.MustCompile("^nplurals=([0-9]+); plural=(.+);$")

func parseMetadata(s string) (Metadata, error) {
	metadata := Metadata{}

	for line := range strings.SplitSeq(s, "\n") {
		key, value, _ := strings.Cut(line, ": ")
		switch key {
		case "Language":
			metadata.Language = value
		case "Plural-Forms":
			// format is "nplurals=%d; plural=n>10|whatever;
			matches := pluralsRe.FindAllStringSubmatch(value, -1)
			if len(matches) != 1 || len(matches[0]) != 3 {
				return metadata, fmt.Errorf("could not parse Plural-Forms")
			}

			nplurals, err := strconv.Atoi(matches[0][1])
			if err != nil {
				return metadata, fmt.Errorf("could not parse 'nplurals=' of Plural-Forms")
			}

			formula, err := parseFormula(matches[0][2])
			if err != nil {
				return metadata, fmt.Errorf("could not parse 'plural=' of Plural-Forms")
			}

			metadata.NPlurals = nplurals
			metadata.Plurals = formula
		}
	}

	return metadata, nil
}
