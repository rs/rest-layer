package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPValidator(t *testing.T) {
	v, err := IP{}.Validate("1.2.3.4")
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3.4", v)
	v, err = IP{}.Validate("2001:1265:0000:0000:0AE4:0000:005B:06B0")
	assert.NoError(t, err)
	assert.Equal(t, "2001:1265::ae4:0:5b:6b0", v)
	v, err = IP{}.Validate(12345)
	assert.EqualError(t, err, "invalid type")
	assert.Equal(t, nil, v)
	v, err = IP{}.Validate("invalid")
	assert.EqualError(t, err, "invalid IP format")
	assert.Equal(t, nil, v)
	v, err = IP{StoreBinary: true}.Validate("1.2.3.4")
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x1, 0x2, 0x3, 0x4}, v)
	v, err = IP{StoreBinary: true}.Validate("2001:1265::ae4:0:5b:6b0")
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x20, 0x1, 0x12, 0x65, 0x0, 0x0, 0x0, 0x0, 0xa, 0xe4, 0x0, 0x0, 0x0, 0x5b, 0x6, 0xb0}, v)
}

func TestIPValidatorSerialize(t *testing.T) {
	v, err := IP{}.Serialize("1.2.3.4")
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3.4", v)
	v, err = IP{StoreBinary: true}.Serialize("1.2.3.4")
	assert.EqualError(t, err, "invalid type")
	assert.Equal(t, nil, v)
	v, err = IP{StoreBinary: true}.Serialize([]byte{0, 1, 2, 3, 4})
	assert.EqualError(t, err, "invalid size")
	assert.Equal(t, nil, v)
	v, err = IP{StoreBinary: true}.Serialize([]byte{1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3.4", v)
}
