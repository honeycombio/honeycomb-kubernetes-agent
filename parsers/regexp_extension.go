package parsers

import "regexp"

// From https://github.com/honeycombio/honeytail/blob/master/parsers/extregexp.go,
// but we don't need to vendor all of honeytail.
// extRegexp is a Regexp with one additional method to make it easier to work
// with named groups
type extRegexp struct {
	*regexp.Regexp
}

func newExtRegexp(expr string) (*extRegexp, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &extRegexp{re}, nil
}

// FindStringSubmatchMap behaves the same as FindStringSubmatch except instead
// of a list of matches with the names separate, it returns the full match and a
// map of named submatches
func (r *extRegexp) FindStringSubmatchMap(s string) (string, map[string]string) {
	match := r.FindStringSubmatch(s)
	if match == nil {
		return "", nil
	}

	captures := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i == 0 {
			continue
		}
		if name != "" {
			// ignore unnamed matches
			captures[name] = match[i]
		}
	}
	return match[0], captures
}
