// +build gofuzz

package query

// FuzzPredicate is used by the https://github.com/dvyukov/go-fuzz framework.
//
// It's method signature must match the prescribed format and it is expected to
// panic upon failure Usage:
//
//     $ go test ./schema/query -update-fuzz-corpus
//     $ go-fuzz-build -func FuzzPredicate -o fuzz-query-predicate.zip github.com/rs/rest-layer/schema/query
//     $ go-fuzz -bin=fuzz-query-predicate.zip -workdir=schema/query/testdata/fuzz-predicate
func FuzzPredicate(data []byte) int {
	_, err := ParsePredicate(string(data))
	if err != nil {
		return 0
	}
	return 1
}

// FuzzProjection is used by the https://github.com/dvyukov/go-fuzz framework.
//
// It's method signature must match the prescribed format and it is expected to
// panic upon failure Usage:
//
//     $ go test ./schema/query -update-fuzz-corpus
//     $ go-fuzz-build -func FuzzProjection -o fuzz-query-projection.zip github.com/rs/rest-layer/schema/query
//     $ go-fuzz -bin=fuzz-query-projection.zip -workdir=schema/query/testdata/fuzz-projection
func FuzzProjection(data []byte) int {
	_, err := ParseProjection(string(data))
	if err != nil {
		return 0
	}
	return 1
}

// FuzzSort is used by the https://github.com/dvyukov/go-fuzz framework.
//
// It's method signature must match the prescribed format and it is expected to
// panic upon failure Usage:
//
//     $ go test ./schema/query -update-fuzz-corpus
//     $ go-fuzz-build -func FuzzSort -o fuzz-query-sort.zip github.com/rs/rest-layer/schema/query
//     $ go-fuzz -bin=fuzz-query-sort.zip -workdir=schema/query/testdata/fuzz-sort
func FuzzSort(data []byte) int {
	_, err := ParseSort(string(data))
	if err != nil {
		return 0
	}
	return 1
}
