package filemerge

import (
	"strings"
)

const (
	markerPrefix = "<!-- jr-stack:"
	markerSuffix = " -->"
	closePrefix  = "<!-- /jr-stack:"
)

// MarkedSectionIDs returns the distinct section IDs that have a well-formed
// jr-stack marker pair (both an opening <!-- jr-stack:ID --> and a matching
// closing <!-- /jr-stack:ID --> marker) in content, in first-seen order.
//
// Malformed markers are ignored — an opening marker with no matching close, an
// ID containing whitespace, or a stray closing marker are never reported. This
// keeps callers that purge by ID from ever acting on a half-written section.
func MarkedSectionIDs(content string) []string {
	var ids []string
	seen := make(map[string]bool)

	offset := 0
	for {
		rel := strings.Index(content[offset:], markerPrefix)
		if rel < 0 {
			break
		}
		idStart := offset + rel + len(markerPrefix)
		end := strings.Index(content[idStart:], markerSuffix)
		if end < 0 {
			break
		}
		id := content[idStart : idStart+end]
		offset = idStart + end + len(markerSuffix)

		if id == "" || strings.ContainsAny(id, " \t\r\n") {
			continue
		}
		if seen[id] {
			continue
		}
		if strings.Contains(content, closeMarker(id)) {
			seen[id] = true
			ids = append(ids, id)
		}
	}

	return ids
}

// openMarker returns the opening marker for a section ID.
func openMarker(sectionID string) string {
	return markerPrefix + sectionID + markerSuffix
}

// closeMarker returns the closing marker for a section ID.
func closeMarker(sectionID string) string {
	return closePrefix + sectionID + markerSuffix
}

// InjectMarkdownSection replaces or appends a marked section in a markdown file.
// Markers use HTML comments: <!-- jr-stack:SECTION_ID --> ... <!-- /jr-stack:SECTION_ID -->
// If the section already exists, its content is replaced.
// If it doesn't exist, it's appended at the end.
// Content outside markers is never touched.
// If content is empty, the section (including markers) is removed.
func InjectMarkdownSection(existing, sectionID, content string) string {
	open := openMarker(sectionID)
	close := closeMarker(sectionID)

	openIdx := strings.Index(existing, open)
	closeIdx := strings.Index(existing, close)

	// If both markers are found and in the correct order, replace the section.
	if openIdx >= 0 && closeIdx >= 0 && closeIdx > openIdx {
		// If content is empty, remove the entire section including markers.
		if content == "" {
			before := existing[:openIdx]
			after := existing[closeIdx+len(close):]

			// Clean up trailing newline after close marker.
			if len(after) > 0 && after[0] == '\n' {
				after = after[1:]
			}
			// Clean up trailing newline before open marker.
			result := strings.TrimRight(before, "\n")
			if after != "" {
				if result != "" {
					result += "\n"
				}
				result += after
			} else if result != "" {
				result += "\n"
			}
			return result
		}

		before := existing[:openIdx]
		after := existing[closeIdx+len(close):]

		var sb strings.Builder
		sb.WriteString(before)
		sb.WriteString(open)
		sb.WriteString("\n")
		sb.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString(close)
		sb.WriteString(after)
		return sb.String()
	}

	// If content is empty and section doesn't exist, return existing unchanged.
	if content == "" {
		return existing
	}

	// Section not found — append at end.
	var sb strings.Builder
	sb.WriteString(existing)
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		sb.WriteString("\n")
	}
	if existing != "" {
		sb.WriteString("\n")
	}
	sb.WriteString(open)
	sb.WriteString("\n")
	sb.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(close)
	sb.WriteString("\n")
	return sb.String()
}
