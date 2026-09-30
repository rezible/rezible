package openapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEmptyInput struct{}

type TestPagination struct {
	Page int `query:"page" default:"1" minimum:"1"`
}

type testInput struct {
	ID        string `path:"id"`
	Selection string `query:"selection"`
	Token     string `header:"X-Test-Token"`
	TestPagination
	Body testInputBody
}

type testInputBody struct {
	Text string `json:"text"`
}

type testOutput struct {
	Body testOutputBody
}

type testOutputBody struct {
	ID        string `json:"id"`
	Selection string `json:"selection"`
	Page      int    `json:"page"`
	Token     string `json:"token"`
	Text      string `json:"text"`
}

func TestOperationRoundTripThroughHuma(t *testing.T) {
	config := huma.DefaultConfig("Test API", "1")
	api := humago.NewWithPrefix(http.NewServeMux(), "/v1", config)
	operation := Operation{
		OperationID:   "publish-report",
		Method:        http.MethodPost,
		Path:          "/reports/{id}",
		DefaultStatus: http.StatusCreated,
	}
	var observedCookies []string
	api.UseMiddleware(func(c huma.Context, next func(huma.Context)) {
		observedCookies = append(observedCookies, c.Header("Cookie"))
		next(c)
	})
	huma.Register(api, operation, func(ctx context.Context, input *testInput) (*testOutput, error) {
		executionCtx := execution.GetContext(ctx)
		assert.True(t, executionCtx.IsAnonymous(), "caller identity must not bypass HTTP authentication")
		assert.Equal(t, execution.SourceHTTP, executionCtx.Provenance.Source)
		_, hasTenant := executionCtx.TenantID()
		assert.False(t, hasTenant)
		return &testOutput{Body: testOutputBody{
			ID:        input.ID,
			Selection: input.Selection,
			Page:      input.Page,
			Token:     input.Token,
			Text:      input.Body.Text,
		}}, nil
	})
	mountedAPI := http.StripPrefix("/api", api.Adapter())
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, execution.SourceInternal, execution.GetContext(r.Context()).Provenance.Source,
			"the test helper must leave HTTP context initialization to the handler")
		ctx := execution.NewRootContext(r.Context(), execution.KindAnonymous, execution.SourceHTTP)
		mountedAPI.ServeHTTP(w, r.WithContext(ctx))
	})
	testAPI := NewTestAPI(t, api, "/v1").WithHandler(handler, "/api/v1")
	aliceAPI := testAPI.WithCookies(&http.Cookie{Name: "session", Value: "alice"})
	bobAPI := testAPI.WithCookies(&http.Cookie{Name: "session", Value: "bob"})
	alicePublish := aliceAPI.Operation[testInput, testOutput](operation)
	bobPublish := bobAPI.Operation[testInput, testOutput](operation)
	input := testInput{
		ID:        "checkout/failed & delayed",
		Selection: "completed & verified",
		Token:     "fixture-token",
		Body:      testInputBody{Text: "Initial report"},
	}
	callerCtx := t.Context()

	first := alicePublish.Call(callerCtx, input)
	input.Body.Text = "Updated report"
	second := bobPublish.Call(callerCtx, input)

	assert.Equal(t, "checkout/failed & delayed", first.Body.ID)
	assert.Equal(t, "completed & verified", first.Body.Selection)
	assert.Equal(t, 1, first.Body.Page, "omitted embedded query field should use Huma's default")
	assert.Equal(t, "fixture-token", first.Body.Token)
	assert.Equal(t, "Initial report", first.Body.Text)
	assert.Equal(t, "Updated report", second.Body.Text)
	assert.NotSame(t, first, second, "each call must decode into a fresh response")
	assert.Equal(t, []string{"session=alice", "session=bob"}, observedCookies)
}

func TestOperationRejectsUnexpectedStatusBeforeDecoding(t *testing.T) {
	_, api := humatest.New(t)
	operation := Operation{
		OperationID: "get-report",
		Method:      http.MethodGet,
		Path:        "/report",
	}
	huma.Register(api, operation, func(context.Context, *testEmptyInput) (*testOutput, error) {
		return nil, huma.Error503ServiceUnavailable("upstream unavailable")
	})
	getReport := NewTestAPI(t, api, "").Operation[testEmptyInput, testOutput](operation)

	response, requestErr := getReport.request(t.Context(), testEmptyInput{})
	require.NoError(t, requestErr)
	statusErr := getReport.checkStatus(response, http.StatusOK)

	require.Error(t, statusErr)
	assert.Contains(t, statusErr.Error(), "get-report: expected HTTP 200, got 503")
	assert.Contains(t, statusErr.Error(), "upstream unavailable")
	getReport.ExpectStatus(t.Context(), testEmptyInput{}, http.StatusServiceUnavailable)
}

func TestOperationReportsMalformedResponse(t *testing.T) {
	_, api := humatest.New(t)
	operation := Operation{
		OperationID: "get-report",
		Method:      http.MethodGet,
		Path:        "/report",
	}
	huma.Register(api, operation, func(context.Context, *testEmptyInput) (*huma.StreamResponse, error) {
		return &huma.StreamResponse{Body: func(c huma.Context) {
			_, writeErr := c.BodyWriter().Write([]byte(`{"text":`))
			assert.NoError(t, writeErr)
		}}, nil
	})
	getReport := NewTestAPI(t, api, "").Operation[testEmptyInput, testOutput](operation)

	response, requestErr := getReport.request(t.Context(), testEmptyInput{})
	require.NoError(t, requestErr)
	require.NoError(t, getReport.checkStatus(response, http.StatusOK))
	output, decodeErr := getReport.decode(response)

	var syntaxErr *json.SyntaxError
	require.ErrorAs(t, decodeErr, &syntaxErr)
	assert.Nil(t, output)
}
