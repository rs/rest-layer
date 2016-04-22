package schema

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeValidate(t *testing.T) {
	now := time.Now().Truncate(time.Minute).UTC()
	for _, f := range formats {
		v, err := Time{}.Validate(now.Format(f))
		assert.NoError(t, err)
		if assert.IsType(t, v, now) {
			assert.True(t, now.Equal(v.(time.Time)), f)
		}
	}
	v, err := Time{}.Validate(now)
	assert.NoError(t, err)
	assert.Equal(t, now, v)
	v, err = Time{}.Validate("invalid date")
	assert.EqualError(t, err, "not a time")
	assert.Nil(t, v)
}
