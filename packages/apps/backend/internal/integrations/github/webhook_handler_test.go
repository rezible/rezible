package github

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOversizedDeliveryIsRejectedBeforeReadingInFull(t *testing.T) {
	body := strings.NewReader(strings.Repeat(" ", 2*maxDeliveryBytes))
	request := httptest.NewRequest(http.MethodPost, "/", body)
	recorder := httptest.NewRecorder()

	newWebhookHandler("secret", nil, nil).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	require.Positive(t, body.Len(), "the handler read the whole body")
}
