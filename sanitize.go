package main

import (
	"regexp"
	"strings"
)

// preserveModelRe matches models whose original spelling must be kept as-is
// (spaces and case intact). Currently: "NIKON D<digits>", e.g. "NIKON D7200".
var preserveModelRe = regexp.MustCompile(`^NIKON D\d+$`)

// sanitizeModel removes characters that are unsafe or ugly in filenames.
func sanitizeModel(s string) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer(
		"\\", "", "/", "", ":", "", "*", "", "?", "", "\"", "",
		"<", "", ">", "", "|", "", " ", "", "\t", "", "\r", "", "\n", "",
	)
	return replacer.Replace(s)
}

// normalizeModel applies cosmetic formatting rules to a sanitized model
// string. Currently: "NIKOND..." -> "NIKON_D...".
func normalizeModel(s string) string {
	if s == "" {
		return s
	}
	return strings.Replace(s, "NIKOND", "NIKON_D", 1)
}

// modelToPrefix turns a raw EXIF Model string into the filename prefix
// (including the trailing underscore), or "" if no prefix should be used.
//
// Rules:
//   - If the raw model matches preserveModelRe exactly ("NIKON D7200"),
//     it is used verbatim, spaces kept: "NIKON D7200_".
//   - Otherwise, the model is sanitized and normalized:
//     "Canon EOS 5D" -> "CanonEOS5D_", "NIKOND850" -> "NIKON_D850_".
func modelToPrefix(rawModel string) string {
	raw := strings.TrimSpace(rawModel)
	if raw == "" {
		return ""
	}
	if preserveModelRe.MatchString(raw) {
		return raw + "_"
	}
	clean := normalizeModel(sanitizeModel(raw))
	if clean == "" {
		return ""
	}
	return clean + "_"
}