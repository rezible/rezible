package hocuspocus

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

type DocumentsService struct {
	serverAddress string
	webhookSecret []byte

	db    *ent.Client
	auth  rez.AuthService
	users rez.UserService
}

func NewDocumentsService(db *ent.Client, auth rez.AuthService, users rez.UserService) (*DocumentsService, error) {
	webhookSecret := rez.Config.GetString("DOCUMENTS_API_SECRET")
	serverAddress := rez.Config.GetString("DOCUMENTS_SERVER_ADDRESS")

	svc := &DocumentsService{
		serverAddress: serverAddress,
		webhookSecret: []byte(webhookSecret),
		db:            db,
		auth:          auth,
		users:         users,
	}

	return svc, nil
}

func (s *DocumentsService) GetServerWebsocketAddress() string {
	return fmt.Sprintf("ws://%s", s.serverAddress)
}

// TODO: these should just be regular api endpoints
func (s *DocumentsService) Handler() http.Handler {
	r := chi.NewRouter()
	r.Post("/auth", s.handleAuthRequest)
	r.Post("/load", s.handleLoadRequest)
	r.Post("/update", s.handleUpdateRequest)
	return r
}

func makeDocumentSessionTokenScopes(docId uuid.UUID) rez.AuthSessionScopes {
	return rez.AuthSessionScopes{
		"documents": []string{docId.String()},
	}
}

func (s *DocumentsService) CreateEditorSessionToken(sess *rez.AuthSession, docId uuid.UUID) (string, error) {
	sess.Scopes = makeDocumentSessionTokenScopes(docId)
	return s.auth.IssueAuthSessionToken(sess)
}

func (s *DocumentsService) verifyRequestSignature(signature []byte, body []byte) bool {
	h := hmac.New(sha256.New, s.webhookSecret)
	h.Write(body)
	digest := []byte(fmt.Sprintf("sha256=%x", h.Sum(nil)))
	return len(signature) == len(digest) && subtle.ConstantTimeCompare(digest, signature) == 1
}

const maxRequestBodyBytes = int64(1024 * 1024)

func (s *DocumentsService) verifyRequestBody(w http.ResponseWriter, r *http.Request, reqBody any) bool {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return false
	}

	body, readErr := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	if readErr != nil {
		log.Error().Err(readErr).Msg("failed to read document request body")
		http.Error(w, readErr.Error(), http.StatusBadRequest)
		return false
	}

	sigHdr := r.Header.Get("X-Rez-Signature-256")
	if sigHdr == "" || !s.verifyRequestSignature([]byte(sigHdr), body) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return false
	}

	if err := json.Unmarshal(body, reqBody); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return false
	}

	return true
}

func (s *DocumentsService) verifyRequestAuth(w http.ResponseWriter, r *http.Request, docId uuid.UUID) context.Context {
	bearerToken := oapi.GetRequestBearerToken(r)
	if bearerToken == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return nil
	}

	sess, verifyErr := s.auth.VerifyAuthSessionToken(bearerToken, makeDocumentSessionTokenScopes(docId))
	if verifyErr != nil {
		http.Error(w, verifyErr.Error(), http.StatusUnauthorized)
		return nil
	}

	authCtx, authErr := s.auth.CreateAuthContext(r.Context(), sess)
	if authErr != nil {
		http.Error(w, authErr.Error(), http.StatusBadRequest)
		return nil
	}

	return authCtx
}

type documentAuthSessionUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type (
	documentAuthRequest struct {
		DocumentId uuid.UUID `json:"documentId"`
	}
	documentAuthResponse struct {
		User     documentAuthSessionUser `json:"user"`
		ReadOnly bool                    `json:"readOnly"`
	}
)

func (s *DocumentsService) handleAuthRequest(w http.ResponseWriter, r *http.Request) {
	var body documentAuthRequest
	if !s.verifyRequestBody(w, r, &body) {
		return
	}
	ctx := s.verifyRequestAuth(w, r, body.DocumentId)
	if ctx == nil {
		return
	}

	usr := s.users.GetUserContext(ctx)

	resp := documentAuthResponse{
		User: documentAuthSessionUser{
			Id:       usr.ID.String(),
			Username: usr.Email,
		},
		ReadOnly: false,
	}
	respBytes, respErr := json.Marshal(resp)
	if respErr != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

type loadRequestBody struct {
	DocumentId uuid.UUID `json:"documentId"`
}

func (s *DocumentsService) handleLoadRequest(w http.ResponseWriter, r *http.Request) {
	var body loadRequestBody
	if !s.verifyRequestBody(w, r, &body) {
		return
	}
	ctx := s.verifyRequestAuth(w, r, body.DocumentId)
	if ctx == nil {
		return
	}

	doc, docErr := s.db.Document.Get(ctx, body.DocumentId)
	if docErr != nil {
		if ent.IsNotFound(docErr) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else {
			log.Error().Err(docErr).Msg("failed to load document")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(doc.Content)
}

type updateRequestBody struct {
	DocumentId uuid.UUID       `json:"documentId"`
	State      json.RawMessage `json:"state"`
}

func (s *DocumentsService) handleUpdateRequest(w http.ResponseWriter, r *http.Request) {
	var body updateRequestBody
	if !s.verifyRequestBody(w, r, &body) {
		return
	}
	ctx := s.verifyRequestAuth(w, r, body.DocumentId)
	if ctx == nil {
		return
	}

	update := s.db.Document.UpdateOneID(body.DocumentId).SetContent(body.State)
	saveErr := update.Exec(ctx)
	if saveErr != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
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

func (s *DocumentsService) parseDocument(raw []byte) (*pmContent, error) {
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

func (s *DocumentsService) ConvertToHTML(ctx context.Context, rawDoc string) (string, error) {
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

func (s *DocumentsService) GetDocumentSchemaSpec(ctx context.Context, schemaName string) (*rez.DocumentSchemaSpec, error) {
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

func (s *DocumentsService) apiRequest(ctx context.Context, endpoint string, method string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("http://%s/api/%s", s.serverAddress, endpoint)
	req, reqErr := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if reqErr != nil {
		return nil, fmt.Errorf("create request: %w", reqErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer foobar")
	resp, doErr := http.DefaultClient.Do(req)
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
