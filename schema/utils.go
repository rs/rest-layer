package schema

import "strings"

// spiltFieldPath splits name on the first dot character and returns the left
// and right sides respectively. The final return parameter indicates weather
// or not the string was split, and helps distinguish the case of name == "foo"
// from name == "foo.".
func splitFieldPath(name string) (string, string, bool) {
	if i := strings.IndexByte(name, '.'); i != -1 {
		remaining := name[i+1:]
		name = name[:i]
		return name, remaining, true
	}
	return name, "", false
}
