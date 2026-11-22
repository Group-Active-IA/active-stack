package jsonedit

import (
	"encoding/json"
	"strings"
	"testing"
)

func mustRoot(t *testing.T, src string) Span {
	t.Helper()
	r, err := RootSpan([]byte(src))
	if err != nil {
		t.Fatalf("RootSpan(%q): %v", src, err)
	}
	return r
}

func TestRootSpanRefusesWhatIsNotAJSONObject(t *testing.T) {
	for name, src := range map[string]string{
		"empty":          "",
		"text":           "hello",
		"array":          "[1]",
		"comment":        "{\n// c\n\"a\": 1}",
		"trailing comma": "{\"a\": 1,}",
		"bom":            "\xef\xbb\xbf{}",
		"unterminated":   "{\"a\": ",
	} {
		if _, err := RootSpan([]byte(src)); err == nil {
			t.Errorf("%s: RootSpan accepted %q", name, src)
		}
	}
}

func TestAddMemberAppendsLikeItsNeighboursAndTouchesNothingElse(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"empty object", "{}", "{\n  \"k\": \"v\"\n}"},
		{"two spaces", "{\n  \"a\": 1\n}\n", "{\n  \"a\": 1,\n  \"k\": \"v\"\n}\n"},
		{"four spaces and CRLF", "{\r\n    \"a\": 1\r\n}\r\n", "{\r\n    \"a\": 1,\r\n    \"k\": \"v\"\r\n}\r\n"},
		{"tabs", "{\n\t\"a\": [1, 2]\n}", "{\n\t\"a\": [1, 2],\n\t\"k\": \"v\"\n}"},
		{"one line", "{\"a\":1}", "{\"a\":1,\n  \"k\": \"v\"}"},
		{"surrounding whitespace", "\n\n{\n  \"a\": 1\n}\n\n", "\n\n{\n  \"a\": 1,\n  \"k\": \"v\"\n}\n\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.src)
			got, err := AddMember(src, mustRoot(t, c.src), "k", "v", DetectStyle(src))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != c.want {
				t.Errorf("got  %q\nwant %q", got, c.want)
			}
			if !json.Valid(got) {
				t.Errorf("result is not valid JSON: %q", got)
			}
		})
	}
}

func TestRemoveMemberIsTheInverseOfAddMember(t *testing.T) {
	for _, src := range []string{
		"{\n  \"a\": 1\n}\n",
		"{\r\n    \"a\": 1,\r\n    \"b\": [1, 2]\r\n}\r\n",
		"{\n\t\"a\": {\"x\": 1}\n}",
	} {
		b := []byte(src)
		root := mustRoot(t, src)
		added, err := AddMember(b, root, "k", "v", DetectStyle(b))
		if err != nil {
			t.Fatal(err)
		}
		ms, _ := ParseObject(added, mustRoot(t, string(added)))
		i, ok := LastMember(ms, "k")
		if !ok {
			t.Fatalf("the added member is not found in %q", added)
		}
		back, err := RemoveMember(added, mustRoot(t, string(added)), i)
		if err != nil {
			t.Fatal(err)
		}
		if string(back) != src {
			t.Errorf("remove(add(x)) != x\n got  %q\n want %q", back, src)
		}
	}
}

func TestArrayElementsAreAddedAndRemovedInPlace(t *testing.T) {
	src := "{\n  \"list\": [\n    1,\n    2\n  ]\n}\n"
	b := []byte(src)
	ms, _ := ParseObject(b, mustRoot(t, src))
	arr := ms[0].Val
	added, err := AddElem(b, arr, "x", DetectStyle(b))
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n  \"list\": [\n    1,\n    2,\n    \"x\"\n  ]\n}\n"; string(added) != want {
		t.Fatalf("got %q want %q", added, want)
	}
	ms, _ = ParseObject(added, mustRoot(t, string(added)))
	es, _ := ParseArray(added, ms[0].Val)
	if len(es) != 3 {
		t.Fatalf("want 3 elements, got %d", len(es))
	}
	back, err := RemoveElem(added, ms[0].Val, 2)
	if err != nil {
		t.Fatal(err)
	}
	if string(back) != src {
		t.Fatalf("got %q want %q", back, src)
	}
}

func TestLastMemberKeepsTheLastDuplicate(t *testing.T) {
	src := `{"a": 1, "b": 2, "a": 3}`
	ms, err := ParseObject([]byte(src), mustRoot(t, src))
	if err != nil {
		t.Fatal(err)
	}
	i, ok := LastMember(ms, "a")
	if !ok || i != 2 {
		t.Fatalf("LastMember = %d, %v; want 2, true", i, ok)
	}
	if _, ok := LastMember(ms, "zzz"); ok {
		t.Fatal("a missing key must not be found")
	}
}

func TestValuesAreNotHTMLEscaped(t *testing.T) {
	src := "{}"
	got, err := AddMember([]byte(src), mustRoot(t, src), "k", "a<b>&c", DetectStyle([]byte(src)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "a<b>&c") {
		t.Fatalf("HTML characters were escaped: %q", got)
	}
}
