package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	rez "github.com/twohundreds/rezible"
	"github.com/twohundreds/rezible/ent"
)

type Service struct {
	serverAddress string
	users         rez.UserService

	httpClient *http.Client
}

func NewService(serverAddress string, users rez.UserService) (*Service, error) {
	svc := &Service{
		serverAddress: serverAddress,
		users:         users,
		httpClient:    &http.Client{},
	}

	return svc, nil
}

//func (s *Service) GetWebhooks() rez.Webhooks {
//	return rez.Webhooks{
//		"documents": http.HandlerFunc(s.webhookHandler),
//	}
//}
//
//func (s *Service) webhookHandler(w http.ResponseWriter, r *http.Request) {
//	body, readErr := io.ReadAll(r.Body)
//	if readErr != nil {
//		log.Error().Err(readErr).Msg("failed to read document webhook body")
//		http.Error(w, readErr.Error(), http.StatusBadRequest)
//		return
//	}
//	log.Debug().Str("documents", "webhook").Msg(string(body))
//	w.WriteHeader(http.StatusOK)
//}

func (s *Service) GetWebsocketAddress() string {
	return fmt.Sprintf("ws://%s", s.serverAddress)
}

func (s *Service) CheckUserDocumentAccess(ctx context.Context, user *ent.User, documentName string) (bool, error) {
	readOnly := false
	if false {
		return false, rez.ErrUnauthorized
	}
	return readOnly, nil
}

type pmMark struct {
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
}

type pmContent struct {
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
	Content    []pmContent    `json:"content"`
	Marks      []pmMark       `json:"marks"`
	Text       string         `json:"text"`
}

func (s *Service) parseDocument(raw []byte) (*pmContent, error) {
	var content pmContent
	if jsonErr := json.Unmarshal(raw, &content); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal json: %w", jsonErr)
	}
	return &content, nil
}

type apiTransformRequest struct {
	Format  string `json:"format"`
	Content string `json:"content"`
}

type apiTransformResponse struct {
	Content string `json:"content"`
}

func (s *Service) ConvertToHTML(ctx context.Context, rawDoc string) (string, error) {
	reqBody, bodyErr := json.Marshal(apiTransformRequest{Format: "html", Content: rawDoc})
	if bodyErr != nil {
		return "", fmt.Errorf("marshal request: %w", bodyErr)
	}
	resp, respErr := s.apiRequest(ctx, "transform", http.MethodPost, reqBody)
	if respErr != nil {
		return "", fmt.Errorf("transform request: %w", respErr)
	}
	var response apiTransformResponse
	if jsonErr := json.Unmarshal(resp, &response); jsonErr != nil {
		return "", fmt.Errorf("unmarshal response: %w", respErr)
	}
	return response.Content, nil
}

type apiSchemaSpecRequest struct {
	Name string `json:"name"`
}

type apiSchemaSpecResponse struct {
	Spec *rez.DocumentSchemaSpec `json:"spec"`
}

func (s *Service) GetDocumentSchemaSpec(ctx context.Context, schemaName string) (*rez.DocumentSchemaSpec, error) {
	reqBody, bodyErr := json.Marshal(apiSchemaSpecRequest{Name: schemaName})
	if bodyErr != nil {
		return nil, fmt.Errorf("marshal request: %w", bodyErr)
	}
	resp, respErr := s.apiRequest(ctx, "schema-spec", http.MethodPost, reqBody)
	if respErr != nil {
		return nil, fmt.Errorf("schema request: %w", respErr)
	}
	var response apiSchemaSpecResponse
	if jsonErr := json.Unmarshal(resp, &response); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal response: %w", respErr)
	}
	return response.Spec, nil
}

func (s *Service) apiRequest(ctx context.Context, endpoint string, method string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("http://%s/api/%s", s.serverAddress, endpoint)
	req, reqErr := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if reqErr != nil {
		return nil, fmt.Errorf("create request: %w", reqErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer foobar")
	resp, doErr := s.httpClient.Do(req)
	if doErr != nil {
		return nil, fmt.Errorf("request: %w", doErr)
	}
	defer func(b io.ReadCloser) {
		if err := b.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close response body")
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response code not 200 %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
