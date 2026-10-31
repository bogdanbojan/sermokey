package epub

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

// buildEPUB assembles a minimal but structurally real EPUB in memory, so
// the tests exercise the actual zip/container/package path rather than a
// convenient stub.
func buildEPUB(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	// The mimetype entry must come first and be stored, not deflated.
	if w, err := zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store}); err == nil {
		w.Write([]byte("application/epub+zip"))
	}
	for name, body := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

const container = `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`

func epub3(t *testing.T) []byte {
	t.Helper()
	return buildEPUB(t, map[string]string{
		"META-INF/container.xml": container,
		"OEBPS/content.opf": `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Le Tour du Monde</dc:title>
    <dc:creator>Jules Verne</dc:creator>
    <dc:language>fr-FR</dc:language>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="c1" href="text/ch1.xhtml" media-type="application/xhtml+xml"/>
    <item id="c2" href="text/ch2.xhtml" media-type="application/xhtml+xml"/>
    <item id="cover" href="images/cover.jpg" media-type="image/jpeg"/>
  </manifest>
  <spine>
    <itemref idref="c1"/>
    <itemref idref="cover"/>
    <itemref idref="c2"/>
  </spine>
</package>`,
		"OEBPS/nav.xhtml": `<html xmlns="http://www.w3.org/1999/xhtml"><body><nav epub:type="toc"><ol>
  <li><a href="text/ch1.xhtml">Le départ</a></li>
  <li><a href="text/ch2.xhtml#start">L'arrivée</a></li>
</ol></nav></body></html>`,
		"OEBPS/text/ch1.xhtml": `<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>ignored</title><style>p{color:red}</style></head>
<body>
  <h1>Chapitre premier</h1>
  <p>Il <em>partit</em> un lundi&nbsp;matin.</p>
  <p>La maison &amp; le jardin.</p>
  <script>var x = 1;</script>
</body></html>`,
		"OEBPS/text/ch2.xhtml": `<html xmlns="http://www.w3.org/1999/xhtml"><body>
  <div><p>Le voyage continue.</p></div>
</body></html>`,
		"OEBPS/images/cover.jpg": "\xff\xd8not-really-a-jpeg",
	})
}

func TestParseEPUB3(t *testing.T) {
	data := epub3(t)
	book, err := Parse(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	if book.Meta.Title != "Le Tour du Monde" || book.Meta.Author != "Jules Verne" {
		t.Errorf("metadata = %q / %q", book.Meta.Title, book.Meta.Author)
	}
	// "fr-FR" must reduce to the primary subtag so it can be matched
	// against the reading provider registry.
	if book.Language != "fr" {
		t.Errorf("language = %q, want fr", book.Language)
	}
	if len(book.Sections) != 2 {
		t.Fatalf("got %d sections, want 2 (the image in the spine must be skipped)", len(book.Sections))
	}

	// Chapter names come from the nav document, including when its href
	// carries a fragment.
	if book.Sections[0].Title != "Le départ" {
		t.Errorf("section 0 title = %q, want %q", book.Sections[0].Title, "Le départ")
	}
	if book.Sections[1].Title != "L'arrivée" {
		t.Errorf("section 1 title = %q, want %q", book.Sections[1].Title, "L'arrivée")
	}
}

func TestParseExtractsReadableProse(t *testing.T) {
	data := epub3(t)
	book, _ := Parse(bytes.NewReader(data), int64(len(data)))
	paras := book.Sections[0].Paragraphs
	if len(paras) != 2 {
		t.Fatalf("got %d paragraphs, want 2: %+v", len(paras), paras)
	}

	// Inline markup must not split a sentence: a reader hovering "partit"
	// needs it inside its sentence, not alone in a paragraph of its own.
	if paras[0].Text != "Il partit un lundi matin." {
		t.Errorf("paragraph 0 = %q", paras[0].Text)
	}
	// Entities have to be decoded, or the tokenizer meets "&amp;".
	if paras[1].Text != "La maison & le jardin." {
		t.Errorf("paragraph 1 = %q", paras[1].Text)
	}
	for _, p := range paras {
		if strings.Contains(p.Text, "var x") || strings.Contains(p.Text, "color:red") {
			t.Errorf("script or style leaked into the text: %q", p.Text)
		}
	}
}

// EPUB 2 is still the bulk of what exists in the wild; its table of
// contents lives in an NCX file referenced from the spine.
func TestParseEPUB2WithNCX(t *testing.T) {
	data := buildEPUB(t, map[string]string{
		"META-INF/container.xml": container,
		"OEBPS/content.opf": `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Contes</dc:title><dc:language>fr</dc:language>
  </metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <item id="c1" href="ch1.html" media-type="application/xhtml+xml"/>
  </manifest>
  <spine toc="ncx"><itemref idref="c1"/></spine>
</package>`,
		"OEBPS/toc.ncx": `<?xml version="1.0"?>
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/"><navMap>
  <navPoint id="n1"><navLabel><text>La chèvre</text></navLabel><content src="ch1.html"/></navPoint>
</navMap></ncx>`,
		"OEBPS/ch1.html": `<html><body><p>Elle broutait l'herbe.</p></body></html>`,
	})
	book, err := Parse(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(book.Sections) != 1 || book.Sections[0].Title != "La chèvre" {
		t.Errorf("sections = %+v", book.Sections)
	}
}

// Without a TOC entry the chapter still needs a name; the first heading in
// the document is the natural one, and it must not then be repeated as the
// opening paragraph.
func TestParseFallsBackToHeadingForTitle(t *testing.T) {
	data := buildEPUB(t, map[string]string{
		"META-INF/container.xml": container,
		"OEBPS/content.opf": `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title>T</dc:title></metadata>
  <manifest><item id="c1" href="a.xhtml" media-type="application/xhtml+xml"/></manifest>
  <spine><itemref idref="c1"/></spine>
</package>`,
		"OEBPS/a.xhtml": `<html><body><h2>Le réveil</h2><p>Il faisait froid.</p></body></html>`,
	})
	book, err := Parse(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if book.Sections[0].Title != "Le réveil" {
		t.Errorf("title = %q", book.Sections[0].Title)
	}
	if len(book.Sections[0].Paragraphs) != 1 {
		t.Errorf("heading was duplicated into the body: %+v", book.Sections[0].Paragraphs)
	}
}

// Manifest hrefs are relative to the package document, not the archive
// root, and are percent-encoded. Getting either wrong means no chapter is
// found and the import silently produces an empty book.
func TestParseResolvesNestedAndEncodedPaths(t *testing.T) {
	data := buildEPUB(t, map[string]string{
		"META-INF/container.xml": `<?xml version="1.0"?>
<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles><rootfile full-path="content/pkg.opf"/></rootfiles></container>`,
		"content/pkg.opf": `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title>T</dc:title></metadata>
  <manifest><item id="c" href="pages/chapitre%20un.xhtml" media-type="application/xhtml+xml"/></manifest>
  <spine><itemref idref="c"/></spine>
</package>`,
		"content/pages/chapitre un.xhtml": `<html><body><p>Trouvé.</p></body></html>`,
	})
	book, err := Parse(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(book.Sections) != 1 || book.Sections[0].Paragraphs[0].Text != "Trouvé." {
		t.Errorf("sections = %+v", book.Sections)
	}
}

func TestParseRejectsInvalidFiles(t *testing.T) {
	for _, tc := range []struct {
		name    string
		data    []byte
		wantSub string
	}{
		{"not a zip", []byte("this is a PDF, honestly"), "invalid zip"},
		{"zip without container", buildEPUB(t, map[string]string{"a.txt": "hello"}), "not a valid EPUB"},
		{
			"no readable text",
			buildEPUB(t, map[string]string{
				"META-INF/container.xml": container,
				"OEBPS/content.opf": `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf"><metadata/>
  <manifest><item id="i" href="cover.jpg" media-type="image/jpeg"/></manifest>
  <spine><itemref idref="i"/></spine></package>`,
				"OEBPS/cover.jpg": "binary",
			}),
			"no readable text",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(bytes.NewReader(tc.data), int64(len(tc.data)))
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("error = %q, want it to mention %q", err, tc.wantSub)
			}
		})
	}
}
