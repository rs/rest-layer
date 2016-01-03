package rest

// compareEtag compares a client provided etag with a base etag. The client provided
// etag may or may not have quotes while the base etag is never quoted. This loose
// comparison of etag allows clients not stricly respecting RFC to send the etag with
// or without quotes when the etag comes from, for instance, the API JSON response.
func compareEtag(etag, baseEtag string) bool {
	if etag == baseEtag {
		return true
	}
	if l := len(etag); l == len(baseEtag)+2 && l > 3 && etag[0] == '"' && etag[l-1] == '"' && etag[1:l-1] == baseEtag {
		return true
	}
	return false
}
