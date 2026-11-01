package reading

import "testing"

// The final row of a generated table used to keep its trailing newline,
// because `line` inferred the row end from the next row's offset and the
// last row has no next offset.
func TestTableLineStripsTrailingNewlineOnLastRow(t *testing.T) {
	tb := newTable("aa\tfirst\nzz\tlast\n")
	if got := tb.line(1); got != "zz\tlast" {
		t.Errorf("line(1) = %q, want %q", got, "zz\tlast")
	}
	rows := tb.lookup("zz", 2)
	if len(rows) != 1 || rows[0][1] != "last" {
		t.Errorf("lookup(zz) = %q, want last field %q", rows, "last")
	}
}

// A blank line used to fold itself into the preceding row.
func TestTableLineSurvivesBlankLines(t *testing.T) {
	tb := newTable("aa\tx\n\nbb\ty\n")
	if got := tb.line(0); got != "aa\tx" {
		t.Errorf("line(0) = %q, want %q", got, "aa\tx")
	}
}
