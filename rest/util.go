package rest

func compareEtag(etag, baseEtag string) bool {
	if etag == baseEtag {
		return true
	}
	if l := len(etag); l == len(baseEtag)+2 && l > 3 && etag[0] == '"' && etag[l-1] == '"' && etag[1:l-1] == baseEtag {
		return true
	}
	return false
}
