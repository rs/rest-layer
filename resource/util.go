package resource

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

// Etag computes an etag based on content of the payload.
func genEtag(payload map[string]interface{}) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(b)), nil
}
