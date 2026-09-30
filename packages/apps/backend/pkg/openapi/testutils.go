package openapi

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/require"
)

// TestAPI calls registered operations through Huma, with typed inputs and outputs.
// The prefix is the adapter's mount path, which is absent from operation paths.
type TestAPI struct {
	t       *testing.T
	api     humatest.TestAPI
	prefix  string
	cookies []*http.Cookie
}

func NewTestAPI(t *testing.T, api API, prefix string) *TestAPI {
	t.Helper()
	return &TestAPI{
		t:      t,
		api:    humatest.Wrap(testAPILog{t}, api),
		prefix: strings.TrimRight(prefix, "/"),
	}
}

// WithHandler dispatches through an enclosing HTTP handler. The prefix includes
// its base path and the API version; operation metadata still comes from the API.
func (a *TestAPI) WithHandler(handler http.Handler, prefix string) *TestAPI {
	a.t.Helper()
	copied := *a
	copied.api = humatest.Wrap(testAPILog{a.t}, testHandlerAPI{API: a.api, handler: handler})
	copied.prefix = strings.TrimRight(prefix, "/")
	return &copied
}

type testHandlerAPI struct {
	API
	handler http.Handler
}

func (a testHandlerAPI) Adapter() Adapter {
	return testHandlerAdapter{Adapter: a.API.Adapter(), handler: a.handler}
}

type testHandlerAdapter struct {
	Adapter
	handler http.Handler
}

func (a testHandlerAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.handler.ServeHTTP(w, r)
}

// Huma dumps whole requests, including authentication cookies. Operation failures
// supply status and body diagnostics instead of logging those request dumps.
type testAPILog struct{ *testing.T }

func (testAPILog) Log(...any)          {}
func (testAPILog) Logf(string, ...any) {}

func (a *TestAPI) WithCookies(cookies ...*http.Cookie) *TestAPI {
	a.t.Helper()
	cp := *a
	cp.cookies = append([]*http.Cookie(nil), cookies...)
	return &cp
}

// Operation binds struct request and response contracts once. Calls use Huma's
// path/query/header tags and JSON Body field; no handler is invoked directly.
func (a *TestAPI) Operation[I, O any](definition Operation) *TestOperation[I, O] {
	a.t.Helper()
	inputType := reflect.TypeFor[I]()
	outputType := reflect.TypeFor[O]()
	require.Equal(a.t, reflect.Struct, inputType.Kind(), "%s request must be a struct", definition.OperationID)
	require.Equal(a.t, reflect.Struct, outputType.Kind(), "%s response must be a struct", definition.OperationID)

	path := a.api.OpenAPI().Paths[definition.Path]
	require.NotNil(a.t, path, "%s is not registered at %s", definition.OperationID, definition.Path)
	methods := map[string]*Operation{
		http.MethodGet:     path.Get,
		http.MethodPost:    path.Post,
		http.MethodPut:     path.Put,
		http.MethodPatch:   path.Patch,
		http.MethodDelete:  path.Delete,
		http.MethodHead:    path.Head,
		http.MethodOptions: path.Options,
		http.MethodTrace:   path.Trace,
	}
	registered := methods[definition.Method]
	require.NotNil(a.t, registered, "%s has no registered %s operation", definition.Path, definition.Method)
	require.Equal(a.t, definition.OperationID, registered.OperationID, "operation identity does not match registered route")
	return &TestOperation[I, O]{api: a, operation: registered}
}

type TestOperation[I, O any] struct {
	api       *TestAPI
	operation *Operation
}

func (o *TestOperation[I, O]) Call(ctx context.Context, input I) *O {
	o.api.t.Helper()
	response, requestErr := o.request(ctx, input)
	require.NoError(o.api.t, requestErr, "%s: build request", o.operation.OperationID)
	require.NoError(o.api.t, o.checkStatus(response, o.operation.DefaultStatus))
	output, decodeErr := o.decode(response)
	require.NoError(o.api.t, decodeErr, "%s: decode response", o.operation.OperationID)
	return output
}

// ExpectStatus checks an error response without decoding it as a success DTO.
func (o *TestOperation[I, O]) ExpectStatus(ctx context.Context, input I, status int) {
	o.api.t.Helper()
	response, requestErr := o.request(ctx, input)
	require.NoError(o.api.t, requestErr, "%s: build request", o.operation.OperationID)
	require.NoError(o.api.t, o.checkStatus(response, status))
}

func (o *TestOperation[I, O]) request(ctx context.Context, input I) (*httptest.ResponseRecorder, error) {
	o.api.t.Helper()
	path := o.operation.Path
	query := url.Values{}
	args := make([]any, 0)
	value := reflect.ValueOf(input)
	for _, field := range reflect.VisibleFields(value.Type()) {
		if field.Anonymous || !field.IsExported() {
			continue
		}
		fieldValue, fieldErr := value.FieldByIndexErr(field.Index)
		if fieldErr != nil {
			return nil, fmt.Errorf("read request field %s: %w", field.Name, fieldErr)
		}
		if field.Name == "Body" {
			if fieldValue.Kind() == reflect.Pointer && fieldValue.IsNil() {
				continue
			}
			args = append(args, fieldValue.Interface())
			continue
		}
		pathName := field.Tag.Get("path")
		queryName := field.Tag.Get("query")
		headerName := field.Tag.Get("header")
		if pathName == "" && queryName == "" && headerName == "" {
			continue
		}
		if pathName == "" && fieldValue.IsZero() {
			continue // Let Huma apply defaults for omitted optional parameters.
		}
		text, textErr := testParameterText(fieldValue)
		if textErr != nil {
			return nil, fmt.Errorf("encode request field %s: %w", field.Name, textErr)
		}
		if pathName != "" {
			path = strings.ReplaceAll(path, "{"+pathName+"}", url.PathEscape(text))
		}
		if queryName != "" {
			query.Set(queryName, text)
		}
		if headerName != "" {
			args = append(args, headerName+": "+text)
		}
	}
	if strings.ContainsAny(path, "{}") {
		return nil, fmt.Errorf("request leaves unresolved path parameters in %s", path)
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	if len(o.api.cookies) > 0 {
		request := &http.Request{Header: http.Header{}}
		for _, cookie := range o.api.cookies {
			request.AddCookie(cookie)
		}
		args = append(args, "Cookie: "+request.Header.Get("Cookie"))
	}
	return o.api.api.DoCtx(ctx, o.operation.Method, o.api.prefix+path, args...), nil
}

func testParameterText(value reflect.Value) (string, error) {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return "", nil
		}
		value = value.Elem()
	}
	if marshaler, ok := value.Interface().(encoding.TextMarshaler); ok {
		text, marshalErr := marshaler.MarshalText()
		return string(text), marshalErr
	}
	switch value.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return fmt.Sprint(value.Interface()), nil
	default:
		return "", fmt.Errorf("unsupported parameter type %s", value.Type())
	}
}

func (o *TestOperation[I, O]) checkStatus(response *httptest.ResponseRecorder, status int) error {
	if response.Code != status {
		return fmt.Errorf("%s: expected HTTP %d, got %d: %.1000s", o.operation.OperationID, status, response.Code, response.Body.Bytes())
	}
	return nil
}

func (o *TestOperation[I, O]) decode(response *httptest.ResponseRecorder) (*O, error) {
	output := new(O)
	body := reflect.ValueOf(output).Elem().FieldByName("Body")
	if !body.IsValid() {
		return output, nil
	}
	if decodeErr := json.Unmarshal(response.Body.Bytes(), body.Addr().Interface()); decodeErr != nil {
		return nil, decodeErr
	}
	return output, nil
}
