package filemerge

import (
	"sort"
	"strings"
)

// RemoveMarkdownSectionsPreservingContent removes all recognized managed marker
// spans for sectionID, preserving every byte outside them. Unmatched markers
// are left in place, including a stray closing marker before a valid section.
func RemoveMarkdownSectionsPreservingContent(existing, sectionID string) string {
	type span struct{ start, end int }
	var spans []span
	for _, pair := range recognizedMarkerPairs() {
		for _, occurrence := range scanSectionOccurrences(existing, pair) {
			if occurrence.id != sectionID {
				continue
			}
			from := occurrence.pos + len(pair.open) + len(sectionID) + len(markerSuffix)
			closing := pair.close + sectionID + markerSuffix
			closeOffset := strings.Index(existing[from:], closing)
			if closeOffset >= 0 {
				spans = append(spans, span{occurrence.pos, from + closeOffset + len(closing)})
			}
		}
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })
	var result strings.Builder
	offset := 0
	for index := 0; index < len(spans); {
		section := spans[index]
		next := index + 1
		for next < len(spans) && spans[next].start < section.end {
			section.end = max(section.end, spans[next].end)
			next++
		}
		result.WriteString(existing[offset:section.start])
		if next > index+1 {
			// Overlapping target spans are ambiguous. Preserve the whole group
			// rather than guessing which opening marker owns the closing one.
			result.WriteString(existing[section.start:section.end])
		}
		offset = section.end
		index = next
	}
	result.WriteString(existing[offset:])
	return result.String()
}
