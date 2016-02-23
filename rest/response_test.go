package rest

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type FakeResponseFormatter struct {
	trace *[]string
}

func (rf FakeResponseFormatter) FormatError(ctx context.Context, headers http.Header, err error, skipBody bool) (context.Context, interface{}) {
	*rf.trace = append(*rf.trace, "SendError")
	return ctx, nil
}

func (rf FakeResponseFormatter) FormatItem(ctx context.Context, headers http.Header, i *resource.Item, skipBody bool) (context.Context, interface{}) {
	*rf.trace = append(*rf.trace, "SendItem")
	return ctx, nil
}

func (rf FakeResponseFormatter) FormatList(ctx context.Context, headers http.Header, l *resource.ItemList, skipBody bool) (context.Context, interface{}) {
	*rf.trace = append(*rf.trace, "SendList")
	return ctx, nil
}

type FakeValidatorSerializer struct {
	err        error
	serialized bool
}

func (v FakeValidatorSerializer) GetField(name string) *schema.Field {
	return nil
}

func (v FakeValidatorSerializer) Prepare(payload map[string]interface{}, original *map[string]interface{}, replace bool) (changes map[string]interface{}, base map[string]interface{}) {
	return nil, nil
}

func (v FakeValidatorSerializer) Validate(changes map[string]interface{}, base map[string]interface{}) (doc map[string]interface{}, errs map[string][]interface{}) {
	return nil, nil
}

func (v *FakeValidatorSerializer) Serialize(payload map[string]interface{}) error {
	v.serialized = true
	return v.err
}

func TestFormatResponse(t *testing.T) {
	var trace []string
	var v *FakeValidatorSerializer
	reset := func() {
		trace = []string{}
		v = &FakeValidatorSerializer{err: nil}
	}
	reset()
	rf := FakeResponseFormatter{trace: &trace}

	_, status, _ := formatResponse(nil, rf, nil, 0, nil, nil, false, nil)
	assert.Equal(t, 0, status)
	assert.Equal(t, []string{}, trace)

	reset()
	_, status, _ = formatResponse(nil, rf, nil, 0, nil, errors.New("test"), false, nil)
	assert.Equal(t, 500, status)
	assert.Equal(t, []string{"SendError"}, trace)

	reset()
	_, status, _ = formatResponse(nil, rf, nil, 0, nil, &resource.Item{}, false, nil)
	assert.Equal(t, 0, status)
	assert.Equal(t, []string{"SendItem"}, trace)

	reset()
	_, status, _ = formatResponse(nil, rf, nil, 0, nil, &resource.Item{}, false, v)
	assert.True(t, v.serialized)
	assert.Equal(t, 0, status)
	assert.Equal(t, []string{"SendItem"}, trace)

	reset()
	v.err = errors.New("test")
	_, status, _ = formatResponse(nil, rf, nil, 0, nil, &resource.Item{}, false, v)
	assert.True(t, v.serialized)
	assert.Equal(t, 500, status)
	assert.Equal(t, []string{"SendError"}, trace)

	reset()
	_, status, _ = formatResponse(nil, rf, nil, 0, nil, &resource.ItemList{Items: []*resource.Item{{}}}, false, nil)
	assert.Equal(t, 0, status)
	assert.Equal(t, []string{"SendList"}, trace)

	reset()
	_, status, _ = formatResponse(nil, rf, nil, 0, nil, &resource.ItemList{Items: []*resource.Item{{}}}, false, v)
	assert.True(t, v.serialized)
	assert.Equal(t, 0, status)
	assert.Equal(t, []string{"SendList"}, trace)

	reset()
	v.err = errors.New("test")
	_, status, _ = formatResponse(nil, rf, nil, 0, nil, &resource.ItemList{Items: []*resource.Item{{}}}, false, v)
	assert.True(t, v.serialized)
	assert.Equal(t, 500, status)
	assert.Equal(t, []string{"SendError"}, trace)
}

func TestDefaultResponseFormatterFormatItem(t *testing.T) {
	rf := DefaultResponseFormatter{}
	ctx := context.Background()
	h := http.Header{}
	rctx, payload := rf.FormatItem(ctx, h, &resource.Item{Payload: map[string]interface{}{"foo": "bar"}}, false)
	assert.Equal(t, http.Header{}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, payload)

	h = http.Header{}
	rctx, payload = rf.FormatItem(ctx, h, &resource.Item{Payload: map[string]interface{}{"foo": "bar"}}, true)
	assert.Equal(t, http.Header{}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, nil, payload)

	h = http.Header{}
	update, _ := time.Parse(time.RFC1123, "Tue, 23 Feb 2016 02:49:16 GMT")
	rctx, payload = rf.FormatItem(ctx, h, &resource.Item{Updated: update}, false)
	assert.Equal(t, http.Header{"Last-Modified": []string{"Tue, 23 Feb 2016 02:49:16 GMT"}}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, map[string]interface{}(nil), payload)

	h = http.Header{}
	rctx, payload = rf.FormatItem(ctx, h, &resource.Item{ETag: "1234"}, false)
	assert.Equal(t, http.Header{"Etag": []string{`"1234"`}}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, map[string]interface{}(nil), payload)
}

func TestDefaultResponseFormatterFormatList(t *testing.T) {
	rf := DefaultResponseFormatter{}
	ctx := context.Background()
	h := http.Header{}
	rctx, payload := rf.FormatList(ctx, h, &resource.ItemList{
		Total: -1,
		Items: []*resource.Item{{Payload: map[string]interface{}{"foo": "bar"}}},
	}, false)
	assert.Equal(t, http.Header{}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, []map[string]interface{}{{"foo": "bar"}}, payload)

	h = http.Header{}
	rctx, payload = rf.FormatList(ctx, h, &resource.ItemList{
		Total: -1,
		Items: []*resource.Item{{Payload: map[string]interface{}{"foo": "bar"}}},
	}, true)
	assert.Equal(t, http.Header{}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, nil, payload)

	h = http.Header{}
	rctx, payload = rf.FormatList(ctx, h, &resource.ItemList{
		Total: 1,
		Page:  1,
		Items: []*resource.Item{{Payload: map[string]interface{}{"foo": "bar"}}},
	}, false)
	assert.Equal(t, http.Header{"X-Total": []string{"1"}, "X-Page": []string{"1"}}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, []map[string]interface{}{{"foo": "bar"}}, payload)

	h = http.Header{}
	rctx, payload = rf.FormatList(ctx, h, &resource.ItemList{
		Total: -1,
		Items: []*resource.Item{{ETag: "123", Payload: map[string]interface{}{"foo": "bar"}}},
	}, false)
	assert.Equal(t, http.Header{}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, []map[string]interface{}{{"foo": "bar", "_etag": "123"}}, payload)
}

func TestDefaultResponseFormatterFormatError(t *testing.T) {
	rf := DefaultResponseFormatter{}
	ctx := context.Background()
	h := http.Header{}
	rctx, payload := rf.FormatError(ctx, h, nil, false)
	assert.Equal(t, http.Header{}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, map[string]interface{}{"message": "Server Error", "code": 500}, payload)

	rctx, payload = rf.FormatError(ctx, h, errors.New("test"), false)
	assert.Equal(t, http.Header{}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, map[string]interface{}{"message": "test", "code": 500}, payload)

	rctx, payload = rf.FormatError(ctx, h, ErrNotFound, false)
	assert.Equal(t, http.Header{}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, map[string]interface{}{"message": "Not Found", "code": 404}, payload)

	rctx, payload = rf.FormatError(ctx, h, &Error{123, "test", map[string][]interface{}{"field": {"error"}}}, false)
	assert.Equal(t, http.Header{}, h)
	assert.Equal(t, rctx, ctx)
	assert.Equal(t, map[string]interface{}{"code": 123, "message": "test", "issues": map[string][]interface{}{"field": []interface{}{"error"}}}, payload)
}
