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