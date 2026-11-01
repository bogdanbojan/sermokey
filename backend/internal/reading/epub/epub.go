// Package epub turns an EPUB file into the reader's document model.
//
// It exists to prove the point the reading package was designed around:
// importing a book is a Catalog concern and nothing else. Nothing in this
// package knows about tokenizers, lexicons, flashcards or HTTP — it
// produces the same DocumentMeta and []Section that the bundled .txt
// corpus produces, and everything downstream is unchanged.
//
// Both EPUB 2 and EPUB 3 are handled, because the difference that matters
// here is small: the spine and manifest are the same, and the table of
// contents (NCX in 2, a nav document in 3) is only consulted for chapter
// titles, which have a perfectly good fallback.
//
// Only text is extracted. Images, styling and embedded fonts are
// deliberately dropped: this is a reading surface for language learning,
// where every word has to be a hoverable token, not a faithful rendering
// of the publisher's layout.
package epub

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

// Limits guard against malformed or hostile files. An EPUB is a ZIP, so
// without them a small upload can expand into an unbounded amount of work.
const (
	// MaxSpineItems caps how many documents are read from the spine.
	MaxSpineItems = 2000
	// MaxTotalRunes caps the total extracted text.
	MaxTotalRunes = 12_000_000
	// maxEntryBytes caps a single decompressed entry.
	maxEntryBytes = 32 << 20
	// MaxZipEntries caps the archive's central directory. zip.NewReader
	// reads it eagerly, so a 32 MB file declaring ~700k entries costs
	// several hundred megabytes of map and a sort before a single byte of
	// content is read.
	MaxZipEntries = 20_000
	// MaxDecompressedBytes caps the *total* decompressed output.
	//
	// The per-entry cap alone is not a zip bomb defence. The spine loop
	// stops on MaxTotalRunes, but that counter only advances for entries
	// that yield paragraphs — so 2000 chapters of <style> padding, which
	// extractText discards, decompressed 64 GB while the counter sat at
	// zero and the break never fired.
	MaxDecompressedBytes = 256 << 20
)

// budget is a shared ceiling on decompressed bytes for one parse.
//
// Reserving and charging are deliberately separate operations. Charging the
// per-entry *allowance* up front — which is what this did at first — spends
// the whole budget in `MaxDecompressedBytes / maxEntryBytes` reads no matter
// how small the entries are. At 256 MB and 32 MB that is eight reads, three
// of which go to container.xml, the package document and the table of
// contents, so every book truncated to about five chapters regardless of
// its actual size. A 40 KB novel of 120 chapters imported six of them.
type budget struct{ remaining int64 }

// reserve returns how many bytes the next read may consume, without
// spending anything. The caller charges what it actually read.
func (b *budget) reserve(max int64) int64 {
	if b.remaining <= 0 {
		return 0
	}
	if max > b.remaining {
		return b.remaining
	}
	return max
}

// charge records bytes that were genuinely decompressed.
func (b *budget) charge(n int64) {
	b.remaining -= n
}

func (b *budget) exhausted() bool { return b.remaining <= 0 }

// Book is a parsed EPUB, ready to be tokenized by reading.Build.
type Book struct {
	Meta     reading.DocumentMeta
	Sections []reading.Section
	// Language is the dc:language declared in the package document,
	// lowercased and reduced to its primary subtag ("fr" from "fr-FR").
	// Empty when the file does not declare one.
	Language string
}

// Parse reads an EPUB from r.
func Parse(r io.ReaderAt, size int64) (*Book, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("not a readable EPUB (invalid zip): %w", err)
	}
	if len(zr.File) > MaxZipEntries {
		return nil, fmt.Errorf("the EPUB declares too many files (%d, limit %d)", len(zr.File), MaxZipEntries)
	}
	files := indexFiles(zr)

	bud := &budget{remaining: MaxDecompressedBytes}

	opfPath, err := rootfilePath(files, bud)
	if err != nil {
		return nil, err
	}
	pkg, err := readPackage(files, opfPath, bud)
	if err != nil {
		return nil, err
	}

	// Paths inside the package document are relative to the package
	// document itself, not to the archive root.
	base := path.Dir(opfPath)

	titles := tocTitles(files, base, pkg, bud)

	book := &Book{Language: primarySubtag(firstNonEmpty(pkg.Metadata.Language...))}
	book.Meta = reading.DocumentMeta{
		Title:   strings.TrimSpace(firstNonEmpty(pkg.Metadata.Title...)),
		Author:  strings.TrimSpace(firstNonEmpty(pkg.Metadata.Creator...)),
		Level:   reading.LevelB2,
		License: "Imported file",
		Source:  "EPUB",
	}
	if book.Meta.Title == "" {
		book.Meta.Title = "Untitled"
	}
	book.Meta.Blurb = blurb(book.Meta.Author)

	hrefByID := map[string]string{}
	mediaByID := map[string]string{}
	for _, item := range pkg.Manifest.Items {
		hrefByID[item.ID] = item.Href
		mediaByID[item.ID] = item.MediaType
	}

	var totalRunes int
	var skipped int
	for i, ref := range pkg.Spine.ItemRefs {
		// The byte budget is checked alongside the rune count, so an
		// archive whose chapters decompress to megabytes of markup that
		// yields no paragraphs still terminates.
		if i >= MaxSpineItems || totalRunes >= MaxTotalRunes || bud.exhausted() {
			break
		}
		href := hrefByID[ref.IDRef]
		if href == "" || !isTextual(mediaByID[ref.IDRef], href) {
			continue
		}
		full := resolve(base, href)
		if full == "" {
			continue
		}
		data, err := readEntry(files, full, bud)
		if err != nil {
			// A missing or unreadable chapter should not fail the import,
			// but a book where most chapters fail is worth noticing rather
			// than silently importing as a two-chapter stub.
			skipped++
			continue
		}

		heading, paragraphs := extractText(data)
		if len(paragraphs) == 0 {
			continue // cover pages, blank separators, image-only plates
		}
		for _, p := range paragraphs {
			totalRunes += len([]rune(p))
		}

		title := firstNonEmpty(titles[full], heading)
		if title == "" {
			title = fmt.Sprintf("%d", len(book.Sections)+1)
		}
		section := reading.Section{Title: title}
		for _, p := range paragraphs {
			section.Paragraphs = append(section.Paragraphs, reading.Paragraph{Text: p})
		}
		book.Sections = append(book.Sections, section)
	}

	if len(book.Sections) == 0 {
		return nil, fmt.Errorf("no readable text found in the EPUB")
	}
	if skipped > 0 {
		log.Printf("epub: %q imported with %d of %d spine items unreadable",
			book.Meta.Title, skipped, len(pkg.Spine.ItemRefs))
	}
	return book, nil
}

func blurb(author string) string {
	if author != "" {
		return "Imported EPUB · " + author
	}
	return "Imported EPUB"
}

// --- Container and package document --------------------------------------

type containerXML struct {
	Rootfiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type packageXML struct {
	Metadata struct {
		Title    []string `xml:"title"`
		Creator  []string `xml:"creator"`
		Language []string `xml:"language"`
	} `xml:"metadata"`
	Manifest struct {
		Items []struct {
			ID         string `xml:"id,attr"`
			Href       string `xml:"href,attr"`
			MediaType  string `xml:"media-type,attr"`
			Properties string `xml:"properties,attr"`
		} `xml:"item"`
	} `xml:"manifest"`
	Spine struct {
		TOC      string `xml:"toc,attr"`
		ItemRefs []struct {
			IDRef  string `xml:"idref,attr"`
			Linear string `xml:"linear,attr"`
		} `xml:"itemref"`
	} `xml:"spine"`
}

func rootfilePath(files map[string]*zip.File, bud *budget) (string, error) {
	data, err := readEntry(files, "META-INF/container.xml", bud)
	if err != nil {
		return "", fmt.Errorf("not a valid EPUB: %w", err)
	}
	var c containerXML
	if err := unmarshal(data, &c); err != nil {
		return "", fmt.Errorf("unreadable META-INF/container.xml: %w", err)
	}
	for _, rf := range c.Rootfiles {
		if rf.FullPath != "" {
			return path.Clean(rf.FullPath), nil
		}
	}
	return "", fmt.Errorf("not a valid EPUB: container.xml names no package document")
}

func readPackage(files map[string]*zip.File, opfPath string, bud *budget) (*packageXML, error) {
	data, err := readEntry(files, opfPath, bud)
	if err != nil {
		return nil, fmt.Errorf("missing package document %s: %w", opfPath, err)
	}
	var pkg packageXML
	if err := unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("unreadable package document: %w", err)
	}
	if len(pkg.Spine.ItemRefs) == 0 {
		return nil, fmt.Errorf("the EPUB has an empty spine")
	}
	return &pkg, nil
}

// unmarshal decodes XML leniently. Real-world EPUBs carry undeclared HTML
// entities and occasionally mis-declared encodings, and a strict parser
// rejects files that every e-reader opens happily.
func unmarshal(data []byte, v any) error {
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	dec.Entity = xml.HTMLEntity
	return dec.Decode(v)
}

// --- Table of contents ----------------------------------------------------

// tocTitles maps a spine document's archive path to its title, from the
// EPUB 3 nav document or the EPUB 2 NCX. The TOC is the only place a
// publisher reliably records human chapter names; a heading inside the
// document is a decent fallback but is often absent or decorative.
func tocTitles(files map[string]*zip.File, base string, pkg *packageXML, bud *budget) map[string]string {
	out := map[string]string{}

	for _, item := range pkg.Manifest.Items {
		if !strings.Contains(item.Properties, "nav") {
			continue
		}
		navPath := resolve(base, item.Href)
		if data, err := readEntry(files, navPath, bud); err == nil {
			for href, label := range navLinks(data) {
				out[resolve(path.Dir(navPath), href)] = label
			}
		}
	}

	if pkg.Spine.TOC != "" {
		for _, item := range pkg.Manifest.Items {
			if item.ID != pkg.Spine.TOC {
				continue
			}
			ncxPath := resolve(base, item.Href)
			data, err := readEntry(files, ncxPath, bud)
			if err != nil {
				break
			}
			var ncx struct {
				Points []struct {
					Label struct {
						Text string `xml:"text"`
					} `xml:"navLabel"`
					Content struct {
						Src string `xml:"src,attr"`
					} `xml:"content"`
				} `xml:"navMap>navPoint"`
			}
			if unmarshal(data, &ncx) != nil {
				break
			}
			for _, p := range ncx.Points {
				if p.Content.Src == "" {
					continue
				}
				key := resolve(path.Dir(ncxPath), p.Content.Src)
				if _, exists := out[key]; !exists {
					out[key] = strings.TrimSpace(p.Label.Text)
				}
			}
		}
	}
	return out
}

// navLinks pulls href/label pairs out of an EPUB 3 nav document.
func navLinks(data []byte) map[string]string {
	out := map[string]string{}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	dec.Entity = xml.HTMLEntity

	var href string
	var text strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if local(t.Name) == "a" {
				href = attr(t, "href")
				text.Reset()
			}
		case xml.CharData:
			if href != "" {
				text.Write(t)
			}
		case xml.EndElement:
			if local(t.Name) == "a" && href != "" {
				if label := strings.Join(strings.Fields(text.String()), " "); label != "" {
					out[stripFragment(href)] = label
				}
				href = ""
			}
		}
	}
	return out
}

// --- Text extraction -------------------------------------------------------

// blockElements end a paragraph. Everything else is inline, so its text
// joins the paragraph being built — which is what keeps a sentence
// wrapped in <em> or <a> from being split into three.
var blockElements = map[string]bool{
	"p": true, "div": true, "br": true, "li": true, "blockquote": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"tr": true, "td": true, "th": true, "section": true, "article": true,
	"figcaption": true, "pre": true, "dd": true, "dt": true, "hr": true,
}

// skipElements have their content discarded entirely.
var skipElements = map[string]bool{
	"script": true, "style": true, "head": true, "svg": true, "title": true,
}

// extractText flattens an XHTML document into paragraphs, and returns the
// first heading it saw so the chapter can be named when the TOC does not
// name it.
func extractText(data []byte) (heading string, paragraphs []string) {
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	dec.Entity = xml.HTMLEntity

	var cur strings.Builder
	var skipDepth int
	var headingDepth int

	flush := func(isHeading bool) {
		text := strings.Join(strings.Fields(cur.String()), " ")
		cur.Reset()
		if text == "" {
			return
		}
		if isHeading && heading == "" {
			heading = text
			// A heading is also the first line of the chapter; keeping it
			// as a paragraph would duplicate it under the section title.
			return
		}
		paragraphs = append(paragraphs, text)
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := local(t.Name)
			if skipDepth > 0 {
				skipDepth++
				continue
			}
			if skipElements[name] {
				skipDepth = 1
				continue
			}
			if blockElements[name] {
				flush(headingDepth > 0)
				if isHeadingTag(name) {
					headingDepth++
				}
			}
		case xml.EndElement:
			name := local(t.Name)
			if skipDepth > 0 {
				skipDepth--
				continue
			}
			if blockElements[name] {
				wasHeading := headingDepth > 0
				flush(wasHeading)
				if isHeadingTag(name) && headingDepth > 0 {
					headingDepth--
				}
			}
		case xml.CharData:
			if skipDepth == 0 {
				cur.Write(t)
			}
		}
	}
	flush(headingDepth > 0)
	return heading, paragraphs
}

func isHeadingTag(name string) bool {
	return len(name) == 2 && name[0] == 'h' && name[1] >= '1' && name[1] <= '6'
}

// --- Zip and path helpers --------------------------------------------------

// indexFiles builds a lookup keyed by cleaned path, plus a lowercase
// fallback: EPUBs produced on case-insensitive filesystems routinely
// disagree with their own manifest about capitalisation.
func indexFiles(zr *zip.Reader) map[string]*zip.File {
	out := make(map[string]*zip.File, len(zr.File)*2)
	names := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		clean := path.Clean(f.Name)
		out[clean] = f
		names = append(names, clean)
	}
	sort.Strings(names)
	for _, n := range names {
		if lower := strings.ToLower(n); lower != n {
			if _, exists := out[lower]; !exists {
				out[lower] = out[n]
			}
		}
	}
	return out
}

// readEntry decompresses one archive member, charged against the shared
// budget so the total across a parse stays bounded.
func readEntry(files map[string]*zip.File, name string, bud *budget) ([]byte, error) {
	name = path.Clean(name)
	f, ok := files[name]
	if !ok {
		f, ok = files[strings.ToLower(name)]
	}
	if !ok {
		return nil, fmt.Errorf("%s not found in the archive", name)
	}
	allowance := bud.reserve(maxEntryBytes)
	if allowance <= 0 {
		return nil, fmt.Errorf("decompression budget exhausted")
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, allowance))
	// Charged even on error: the bytes were decompressed either way, which
	// is the cost the budget exists to bound.
	bud.charge(int64(len(data)))
	return data, err
}

// resolve joins a manifest href to the directory holding the document that
// referenced it, and refuses to escape the archive root.
func resolve(base, href string) string {
	href = stripFragment(href)
	if strings.HasPrefix(href, "/") {
		return path.Clean(strings.TrimPrefix(href, "/"))
	}
	joined := path.Clean(path.Join(base, href))
	// TrimPrefix stripped a single "../", so "../../../x" still escaped.
	// Nothing from the archive is written to disk today, but this is only
	// ever used as a lookup key by accident of the current design.
	if !fs.ValidPath(joined) {
		return ""
	}
	return joined
}

func stripFragment(href string) string {
	if i := strings.IndexByte(href, '#'); i >= 0 {
		href = href[:i]
	}
	return unescapePath(href)
}

// unescapePath decodes the percent-encoding EPUBs use for spaces and
// accented characters in filenames.
func unescapePath(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			if hi, ok1 := hexVal(s[i+1]); ok1 {
				if lo, ok2 := hexVal(s[i+2]); ok2 {
					b.WriteByte(hi<<4 | lo)
					i += 2
					continue
				}
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func hexVal(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

func isTextual(mediaType, href string) bool {
	if mediaType != "" {
		return strings.Contains(mediaType, "xhtml") || strings.Contains(mediaType, "html")
	}
	ext := strings.ToLower(path.Ext(href))
	return ext == ".xhtml" || ext == ".html" || ext == ".htm"
}

func local(n xml.Name) string { return strings.ToLower(n.Local) }

func attr(e xml.StartElement, name string) string {
	for _, a := range e.Attr {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func primarySubtag(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if i := strings.IndexAny(lang, "-_"); i > 0 {
		return lang[:i]
	}
	return lang
}
