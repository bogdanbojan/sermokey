package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"testing"
)

// buildEPUB produces a minimal but valid EPUB with n chapters.
func buildNovel(t *testing.T, chapters int, paragraphsEach int) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	add := func(name, body string) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}

	add("mimetype", "application/epub+zip")
	add("META-INF/container.xml", `<?xml version="1.0"?>
<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container" version="1.0">
  <rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`)

	var manifest, spine bytes.Buffer
	for i := 0; i < chapters; i++ {
		id := fmt.Sprintf("ch%d", i)
		fmt.Fprintf(&manifest, `<item id="%s" href="%s.xhtml" media-type="application/xhtml+xml"/>`, id, id)
		fmt.Fprintf(&spine, `<itemref idref="%s"/>`, id)

		var body bytes.Buffer
		fmt.Fprintf(&body, `<?xml version="1.0" encoding="utf-8"?><html xmlns="http://www.w3.org/1999/xhtml"><body><h1>Chapitre %d</h1>`, i+1)
		for p := 0; p < paragraphsEach; p++ {
			fmt.Fprintf(&body, `<p>Le marché du samedi est ouvert et la foule remplit la place numéro %d.</p>`, p)
		}
		body.WriteString(`</body></html>`)
		add("OEBPS/"+id+".xhtml", body.String())
	}

	add("OEBPS/content.opf", fmt.Sprintf(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Un long roman</dc:title><dc:creator>Anon</dc:creator><dc:language>fr</dc:language>
  </metadata>
  <manifest>%s</manifest>
  <spine>%s</spine>
</package>`, manifest.String(), spine.String()))

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// A full-length book must import in full. The decompression budget charged
// each entry its whole 32 MB *allowance* rather than the bytes it actually
// read, so the 256 MB ceiling was spent after eight reads — three of which
// go to container.xml, the package document and the table of contents.
// Every book, however small, truncated to about five chapters.
func TestFullBookImportsEveryChapter(t *testing.T) {
	const chapters = 120
	data := buildNovel(t, chapters, 12)
	t.Logf("archive is %d KB for %d chapters", len(data)/1024, chapters)

	book, err := Parse(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(book.Sections) != chapters {
		t.Fatalf("imported %d of %d chapters", len(book.Sections), chapters)
	}
	// Spot-check the ends, so a truncation in the middle is caught too.
	if got := book.Sections[0].Title; got != "Chapitre 1" {
		t.Errorf("first section title = %q", got)
	}
	if got := book.Sections[chapters-1].Title; got != fmt.Sprintf("Chapitre %d", chapters) {
		t.Errorf("last section title = %q", got)
	}
}

// The budget still has to stop a genuinely hostile archive.
func TestBudgetStopsOversizedContent(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name string, body []byte) {
		w, _ := zw.Create(name)
		_, _ = w.Write(body)
	}
	add("META-INF/container.xml", []byte(`<?xml version="1.0"?>
<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container" version="1.0">
  <rootfiles><rootfile full-path="content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`))

	// Chapters of pure padding that yields no paragraphs: the rune counter
	// never advances, so only the byte budget can stop this.
	padding := bytes.Repeat([]byte("<style>a{color:red}</style>"), 400_000)
	var manifest, spine bytes.Buffer
	const n = 400
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("c%d", i)
		fmt.Fprintf(&manifest, `<item id="%s" href="%s.xhtml" media-type="application/xhtml+xml"/>`, id, id)
		fmt.Fprintf(&spine, `<itemref idref="%s"/>`, id)
		body := append([]byte(`<html><body>`), padding...)
		add(id+".xhtml", append(body, []byte(`</body></html>`)...))
	}
	add("content.opf", []byte(fmt.Sprintf(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0"><metadata xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title>Bomb</dc:title></metadata>
<manifest>%s</manifest><spine>%s</spine></package>`, manifest.String(), spine.String())))
	_ = zw.Close()

	data := buf.Bytes()
	t.Logf("bomb compresses to %d KB, expands to ~%d MB",
		len(data)/1024, n*len(padding)/(1<<20))

	// No paragraphs anywhere, so this must fail rather than hang.
	if _, err := Parse(bytes.NewReader(data), int64(len(data))); err == nil {
		t.Fatal("expected the padded archive to be rejected")
	}
}

// A novel-sized book: ~360 chapters of ~2,500 words, roughly War and Peace.
// Guards the rune and spine ceilings as well as the byte budget.
func TestNovelSizedBookImportsInFull(t *testing.T) {
	const chapters = 360
	data := buildNovel(t, chapters, 180)

	book, err := Parse(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	runes := 0
	for _, s := range book.Sections {
		for _, p := range s.Paragraphs {
			runes += len([]rune(p.Text))
		}
	}
	t.Logf("%d KB archive -> %d sections, %d runes", len(data)/1024, len(book.Sections), runes)

	if len(book.Sections) != chapters {
		t.Fatalf("imported %d of %d chapters", len(book.Sections), chapters)
	}
}
