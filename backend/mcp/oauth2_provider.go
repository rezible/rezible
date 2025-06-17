package mcp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/go-chi/cors"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ProtectedResourceMetadata struct {
	Resource             string   `json:"resource"`
	AuthorizationServers []string `json:"authorization_servers"`
}

type AuthorizationServerMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                   []string `json:"scopes_supported,omitempty"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`
}

type OAuth2Provider struct {
	store    fosite.Storage
	provider fosite.OAuth2Provider
	issuer   string
	prefix   string

	authServerMetadata        []byte
	protectedResourceMetadata []byte
}

func NewOAuth2Provider(prefix string) (*OAuth2Provider, error) {
	p := &OAuth2Provider{
		issuer: "http://localhost:8888",
		prefix: prefix,
	}

	if provErr := p.makeFositeProvider(); provErr != nil {
		return nil, provErr
	}

	if mdErr := p.makeMetadata(); mdErr != nil {
		return nil, mdErr
	}

	return p, nil
}

func (p *OAuth2Provider) makeFositeProvider() error {
	p.store = storage.NewExampleStore()
	secret := []byte("my super secret signing password")

	privateKey, keyErr := rsa.GenerateKey(rand.Reader, 2048)
	if keyErr != nil {
		return fmt.Errorf("failed to generate private key: %w", keyErr)
	}

	config := &fosite.Config{
		AccessTokenLifespan:          time.Minute * 30,
		GlobalSecret:                 secret,
		AccessTokenIssuer:            p.issuer,
		IDTokenIssuer:                p.issuer,
		ClientAuthenticationStrategy: p.verifyClient,
		// ...
	}
	p.provider = compose.ComposeAllEnabled(config, p.store, privateKey)
	return nil
}

func (p *OAuth2Provider) verifyClient(ctx context.Context, request *http.Request, values url.Values) (fosite.Client, error) {
	log.Debug().Msg("verify client")
	return nil, fosite.ErrInvalidClient
}

func (p *OAuth2Provider) makeMetadata() error {
	routesUrl := p.issuer + p.prefix
	authServer := AuthorizationServerMetadata{
		Issuer:                            p.issuer,
		AuthorizationEndpoint:             routesUrl + "/auth",
		TokenEndpoint:                     routesUrl + "/token",
		RegistrationEndpoint:              routesUrl + "/register",
		ResponseTypesSupported:            []string{"code"},
		GrantTypesSupported:               []string{"authorization_code", "client_credentials"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		ScopesSupported:                   []string{"read", "write", "mcp:tools"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_basic", "client_secret_post", "none"},
	}

	protectedResource := ProtectedResourceMetadata{
		Resource:             p.issuer,
		AuthorizationServers: []string{p.issuer},
	}

	var jsonErr error
	p.authServerMetadata, jsonErr = json.MarshalIndent(authServer, "", "  ")
	if jsonErr != nil {
		return fmt.Errorf("failed to marshal authorization server metadata: %w", jsonErr)
	}
	p.protectedResourceMetadata, jsonErr = json.MarshalIndent(protectedResource, "", "  ")
	if jsonErr != nil {
		return fmt.Errorf("failed to marshal protected resource metadata: %w", jsonErr)
	}
	return nil
}

func (p *OAuth2Provider) MakeHandler() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*", "http://localhost:6274"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},

		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.HandleFunc("/auth", p.authEndpoint)
	router.HandleFunc("/token", p.tokenEndpoint)
	router.Post("/register", p.registerClientEndpoint)
	router.HandleFunc("/revoke", p.revokeEndpoint)
	router.HandleFunc("/introspect", p.introspectionEndpoint)

	return router
}

func (p *OAuth2Provider) AuthorizationServerMetadataDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(p.authServerMetadata)
}

func (p *OAuth2Provider) ProtectedResourceMetadataDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(p.protectedResourceMetadata)
}

func (p *OAuth2Provider) newSession(user string) *openid.DefaultSession {
	return &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:      p.issuer,
			Subject:     user,
			Audience:    []string{p.issuer},
			ExpiresAt:   time.Now().Add(time.Hour * 6),
			IssuedAt:    time.Now(),
			RequestedAt: time.Now(),
			AuthTime:    time.Now(),
		},
		Headers: &jwt.Headers{
			Extra: make(map[string]interface{}),
		},
	}
}

type ClientRegistrationRequest struct {
	RedirectUris            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	ClientName              string   `json:"client_name,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
}

type ClientRegistrationResponse struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret,omitempty"`
	RedirectUris            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	ClientName              string   `json:"client_name,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	Scope                   string   `json:"scope,omitempty"`
	ClientSecretExpiresAt   int64    `json:"client_secret_expires_at"`
}

type RFC7591ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func writeErrorResponse(w http.ResponseWriter, errorCode, description string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(statusCode)

	errorResp := RFC7591ErrorResponse{
		Error:            errorCode,
		ErrorDescription: description,
	}

	json.NewEncoder(w).Encode(errorResp)
}

func (p *OAuth2Provider) generateClientID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("mcp_client_%x", bytes), nil
}

func (p *OAuth2Provider) generateClientSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

func (p *OAuth2Provider) registerClientEndpoint(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req ClientRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "invalid_request", "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	// Validate redirect URIs (required field)
	if len(req.RedirectUris) == 0 {
		writeErrorResponse(w, "invalid_redirect_uri", "redirect_uris is required", http.StatusBadRequest)
		return
	}

	// Validate redirect URIs meet MCP security requirements
	/*
		for _, uri := range req.RedirectUris {
			if err := validateRedirectURI(uri); err != nil {
				writeErrorResponse(w, "invalid_redirect_uri", err.Error(), http.StatusBadRequest)
				return
			}
		}
	*/

	// Set MCP-compliant defaults
	if len(req.GrantTypes) == 0 {
		req.GrantTypes = []string{"authorization_code"}
	}
	if len(req.ResponseTypes) == 0 {
		req.ResponseTypes = []string{"code"}
	}
	if req.TokenEndpointAuthMethod == "" {
		// Default to public client for MCP (no client secret)
		req.TokenEndpointAuthMethod = "none"
	}

	// Validate grant types and response types consistency
	/*
		if err := validateGrantAndResponseTypes(req.GrantTypes, req.ResponseTypes); err != nil {
			writeErrorResponse(w, "invalid_request", err.Error(), http.StatusBadRequest)
			return
		}
	*/

	// Generate client ID
	clientID, idErr := p.generateClientID()
	if idErr != nil {
		writeErrorResponse(w, "server_error", "Failed to generate client ID", http.StatusInternalServerError)
		return
	}

	// Generate client secret only for confidential clients
	var clientSecret string
	var secretExpiresAt int64
	isPublic := req.TokenEndpointAuthMethod == "none"

	if !isPublic {
		sec, secretErr := p.generateClientSecret()
		if secretErr != nil {
			writeErrorResponse(w, "server_error", "Failed to generate client secret", http.StatusInternalServerError)
			return
		}
		clientSecret = sec
		secretExpiresAt = 0
	}

	/*
		client := &fosite.DefaultClient{
			ID:            clientID,
			Secret:        []byte(clientSecret), // Hash this in production
			RedirectURIs:  req.RedirectUris,
			GrantTypes:    req.GrantTypes,
			ResponseTypes: req.ResponseTypes,
			Scopes:        strings.Fields(req.Scope),
			Public:        isPublic,
			//Name:                    req.ClientName,
			//TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		}
	*/
	// set client in store

	// Prepare response
	response := ClientRegistrationResponse{
		ClientID:                clientID,
		RedirectUris:            req.RedirectUris,
		GrantTypes:              req.GrantTypes,
		ResponseTypes:           req.ResponseTypes,
		ClientName:              req.ClientName,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		Scope:                   req.Scope,
		ClientSecretExpiresAt:   secretExpiresAt,
	}

	if !isPublic {
		response.ClientSecret = clientSecret
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (p *OAuth2Provider) authEndpoint(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	ar, reqErr := p.provider.NewAuthorizeRequest(ctx, req)
	if reqErr != nil {
		log.Error().Err(reqErr).Msg("Error occurred in NewAuthorizeRequest")
		p.provider.WriteAuthorizeError(ctx, rw, ar, reqErr)
		return
	}

	var requestedScopes string
	for _, this := range ar.GetRequestedScopes() {
		requestedScopes += fmt.Sprintf(`<li><input type="checkbox" name="scopes" value="%s">%s</li>`, this, this)
	}

	req.ParseForm()
	if req.PostForm.Get("username") != "peter" {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.Write([]byte(`<h1>Login page</h1>`))
		rw.Write([]byte(fmt.Sprintf(`
			<p>Howdy! This is the log in page. For this example, it is enough to supply the username.</p>
			<form method="post">
				<p>
					By logging in, you consent to grant these scopes:
					<ul>%s</ul>
				</p>
				<input type="text" name="username" /> <small>try peter</small><br>
				<input type="submit">
			</form>
		`, requestedScopes)))
		return
	}

	for _, scope := range req.PostForm["scopes"] {
		ar.GrantScope(scope)
	}

	mySessionData := p.newSession("peter")

	// When using the HMACSHA strategy you must use something that implements the HMACSessionContainer.
	// It brings you the power of overriding the default values.
	//
	// mySessionData.HMACSession = &strategy.HMACSession{
	//	AccessTokenExpiry: time.Now().Add(time.Day),
	//	AuthorizeCodeExpiry: time.Now().Add(time.Day),
	// }
	//

	// If you're using the JWT strategy, there's currently no distinction between access token and authorize code claims.
	// Therefore, you both access token and authorize code will have the same "exp" claim. If this is something you
	// need let us know on github.
	//
	// mySessionData.JWTClaims.ExpiresAt = time.Now().Add(time.Day)

	// It's also wise to check the requested scopes, e.g.:
	// if ar.GetRequestedScopes().Has("admin") {
	//     http.Error(rw, "you're not allowed to do that", http.StatusForbidden)
	//     return
	// }

	response, respErr := p.provider.NewAuthorizeResponse(ctx, ar, mySessionData)
	if respErr != nil {
		log.Error().Err(respErr).Msg("Error occurred in NewAuthorizeResponse")
		p.provider.WriteAuthorizeError(ctx, rw, ar, respErr)
		return
	}

	p.provider.WriteAuthorizeResponse(ctx, rw, ar, response)
}

func (p *OAuth2Provider) tokenEndpoint(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// Create an empty session object which will be passed to the request handlers
	mySessionData := p.newSession("")

	accessRequest, reqErr := p.provider.NewAccessRequest(ctx, req, mySessionData)
	if reqErr != nil {
		log.Error().Err(reqErr).Msg("Error occurred in NewAccessRequest")
		p.provider.WriteAccessError(ctx, rw, accessRequest, reqErr)
		return
	}

	// If this is a client_credentials grant, grant all requested scopes
	// NewAccessRequest validated that all requested scopes the client is allowed to perform
	// based on configured scope matching strategy.
	if accessRequest.GetGrantTypes().ExactOne("client_credentials") {
		for _, scope := range accessRequest.GetRequestedScopes() {
			accessRequest.GrantScope(scope)
		}
	}

	response, respErr := p.provider.NewAccessResponse(ctx, accessRequest)
	if respErr != nil {
		log.Error().Err(respErr).Msg("Error occurred in NewAccessResponse")
		p.provider.WriteAccessError(ctx, rw, accessRequest, respErr)
		return
	}

	p.provider.WriteAccessResponse(ctx, rw, accessRequest, response)
}

func (p *OAuth2Provider) revokeEndpoint(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	reqErr := p.provider.NewRevocationRequest(ctx, req)
	p.provider.WriteRevocationResponse(ctx, rw, reqErr)
}

func (p *OAuth2Provider) introspectionEndpoint(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	mySessionData := p.newSession("")
	ir, reqErr := p.provider.NewIntrospectionRequest(ctx, req, mySessionData)
	if reqErr != nil {
		log.Error().Err(reqErr).Msg("Error occurred in NewIntrospectionRequest")
		p.provider.WriteIntrospectionError(ctx, rw, reqErr)
		return
	}
	p.provider.WriteIntrospectionResponse(ctx, rw, ir)
}
