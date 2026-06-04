package catalog

// ParseForTest parses a raw YAML catalog for use in tests that need a custom
// catalog (e.g. the hermetic E2E that points harnesses at file:// fixture repos).
//
// This is intentionally a production-package function (not _test.go) so that
// external test packages (e.g. e2e/starter) can call it without importing an
// internal test file. It is the minimal seam needed for hermetic E2E fixtures.
//
// The full validation logic runs — a malformed YAML returns an error. This
// ensures hermetic tests exercise the same catalog rules as production.
func ParseForTest(data []byte) (*Catalog, error) {
	return parse(data)
}
