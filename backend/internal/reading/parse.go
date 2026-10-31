package reading

import (
	"strconv"
	"strings"
)

// ParseDocument reads the bundled text format: a block of `key: value`
// headers, a blank line, then the body. Inside the body, a line beginning
// with "# " starts a new section and blank lines separate paragraphs.
//
// The format is deliberately plain rather than YAML or JSON. It keeps the
// texts readable and diffable as prose, it needs no dependency (the backend
// has none), and it matches how the vocabulary CSV already works: adding
// content is a data change, never a code change.
//
// Parse errors panic because this runs at package init against files we
// author and embed. A malformed text is a build-time bug, exactly as a
// malformed row in vocab_data.csv is.
func ParseDocument(filename, content string) (DocumentMeta, []Section) {
	fail := func(msg string) {
		panic("reading: " + filename + ": " + msg)
	}

	header, body, found := strings.Cut(normalizeNewlines(content), "\n\n")
	if !found {
		fail("no blank line separating the header from the body")
	}

	meta := DocumentMeta{}
	for _, line := range strings.Split(header, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			fail("header line is not `key: value`: " + line)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch key {
		case "id":
			meta.ID = value
		case "title":
			meta.Title = value
		case "author":
			meta.Author = value
		case "level":
			meta.Level = Level(value)
		case "blurb":
			meta.Blurb = value
		case "source":
			meta.Source = value
		case "license":
			meta.License = value
		default:
			fail("unknown header key " + strconv.Quote(key))
		}
	}
	for field, value := range map[string]string{
		"id": meta.ID, "title": meta.Title, "license": meta.License, "blurb": meta.Blurb,
	} {
		if value == "" {
			fail("missing required header " + strconv.Quote(field))
		}
	}
	switch meta.Level {
	case LevelA2, LevelB1, LevelB2:
	default:
		fail("level must be A2, B1 or B2, got " + strconv.Quote(string(meta.Level)))
	}

	var sections []Section
	for _, block := range strings.Split(body, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		if title, ok := strings.CutPrefix(block, "# "); ok {
			sections = append(sections, Section{Title: strings.TrimSpace(title)})
			continue
		}
		if len(sections) == 0 {
			fail("body starts with a paragraph before any `# ` section heading")
		}
		// Join wrapped lines back into one paragraph so the source file can
		// be hard-wrapped for readability without that affecting rendering.
		text := strings.Join(strings.Fields(strings.ReplaceAll(block, "\n", " ")), " ")
		last := &sections[len(sections)-1]
		last.Paragraphs = append(last.Paragraphs, Paragraph{Text: text})
	}
	if len(sections) == 0 {
		fail("no sections")
	}
	return meta, sections
}

func normalizeNewlines(s string) string {
	if !strings.Contains(s, "\r") {
		return s
	}
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}
