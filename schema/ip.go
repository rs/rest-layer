package schema

import (
	"errors"
	"net"
)

// IP validates IP values
type IP struct {
	// StoreBinary activates storage of the IP as binary to save space.
	// The storage requirement is 4 bytes for IPv4 and 16 bytes for IPv6.
	StoreBinary bool
}

// Validate implements FieldValidator
func (v IP) Validate(value interface{}) (interface{}, error) {
	s, ok := value.(string)
	if !ok {
		return nil, errors.New("invalid type")
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, errors.New("invalid IP format")
	}
	if v.StoreBinary {
		// If IP is a v4, store it's 4 bytes representation to save space.
		if v4 := ip.To4(); v4 != nil {
			return []byte(v4), nil
		}
		return []byte(ip), nil
	}
	return ip.String(), nil
}

// Serialize implements FieldSerializer.
func (v IP) Serialize(value interface{}) (interface{}, error) {
	if !v.StoreBinary {
		return value, nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil, errors.New("invalid type")
	}
	if len(b) != 4 && len(b) != 16 {
		return nil, errors.New("invalid size")
	}
	return net.IP(b).String(), nil
}
