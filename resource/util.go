package resource

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	"github.com/rs/rest-layer/schema/query"
)

// Etag computes an etag based on containt of the payload.
func genEtag(payload map[string]interface{}) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(b)), nil
}

func valuesToInterface(v []query.Value) []interface{} {
	I := make([]interface{}, len(v))
	for i, _v := range v {
		I[i] = _v
	}
	return I
}
