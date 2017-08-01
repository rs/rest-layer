package query

import "flag"

// updateFuzzCorpus is set to true when tests are expected to update their fuzz
// test corpus.
//
// Use go test ./schema/query -update-fuzz-corpus to update fuzz test corpus.
var updateFuzzCorpus = flag.Bool("update-fuzz-corpus", false, "update testdata/fuzz-*/corpus/ files")
