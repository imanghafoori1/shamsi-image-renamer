package main

import "testing"

func TestSanitizeModel(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Canon EOS 5D", "CanonEOS5D"},
		{"NIKON D7200", "NIKOND7200"},
		{"  Foo/Bar:Baz*Qux?  ", "FooBarBazQux"},
		{"", ""},
		{"<a>|<b>", "ab"},
		{"with\ttab\nnewline", "withtabnewline"},
		{`quote"and'single`, "quoteand'single"}, // apostrophe is legal in filenames on Windows
	}
	for _, c := range cases {
		if got := sanitizeModel(c.in); got != c.want {
			t.Errorf("sanitizeModel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeModel(t *testing.T) {
	cases := []struct{ in, want string }{
		{"NIKOND7200", "NIKON_D7200"},
		{"NIKOND850", "NIKON_D850"},
		{"NIKOND", "NIKON_D"},
		{"CanonEOS5D", "CanonEOS5D"},
		{"", ""},
		{"nikond7200", "nikond7200"},
	}
	for _, c := range cases {
		if got := normalizeModel(c.in); got != c.want {
			t.Errorf("normalizeModel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestModelToPrefix(t *testing.T) {
	cases := []struct{ in, want string }{
		// Preserve rule: exact "NIKON D<digits>" keeps spaces and case.
		{"NIKON D7200", "NIKON D7200_"},
		{"NIKON D850", "NIKON D850_"},
		{"  NIKON D7200  ", "NIKON D7200_"},

		// Non-matching Nikon models go through sanitize + normalize.
		{"NIKOND7200", "NIKON_D7200_"},
		{"NIKON D7200 II", "NIKON_D7200II_"},

		// Other brands untouched by normalization.
		{"Canon EOS 5D", "CanonEOS5D_"},
		{"Canon EOS 5D Mark IV", "CanonEOS5DMarkIV_"},

		// Empty / whitespace-only.
		{"", ""},
		{"   ", ""},
	}
	for _, c := range cases {
		if got := modelToPrefix(c.in); got != c.want {
			t.Errorf("modelToPrefix(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}