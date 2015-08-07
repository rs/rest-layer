package rest

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestContextError(t *testing.T) {
	assert.Equal(t, ClientClosedRequestError, ContextError(context.Canceled))
	assert.Equal(t, GatewayTimeoutError, ContextError(context.DeadlineExceeded))
	assert.Nil(t, ContextError(nil))
	assert.Equal(t, &Error{520, "test", nil}, ContextError(errors.New("test")))
}

func TestError(t *testing.T) {
	e := &Error{123, "message", nil}
	assert.Equal(t, "message", e.Error())
}
