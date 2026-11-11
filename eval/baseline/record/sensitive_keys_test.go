package record

import (
	"strings"
	"testing"
)

func TestRedactJSONRemovesSensitiveKeys(t *testing.T) {
	got, _ := RedactJSON([]byte(`{"person@example.com":{"count":1},"safe":true}`))
	if strings.Contains(string(got), "person@example.com") || !strings.Contains(string(got), `"safe":true`) {
		t.Fatalf("unexpected shared JSON: %s", got)
	}
}
