package strings

import "strings"

// Join concatenates a variable number of strings.
func Join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

// JoinWithSeparator concatenates a variable number of strings using a separator
// between each one of the elements.
func JoinWithSeparator(sep string, strs ...string) string {
	if len(strs) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, str := range strs[:len(strs)-1] {
		sb.WriteString(str)
		sb.WriteString(sep)
	}
	sb.WriteString(strs[len(strs)-1])

	return sb.String()
}
