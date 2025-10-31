package scim

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

func makeScimHandler(mw chi.Middlewares) (http.Handler, error) {
	if true {
		return http.NewServeMux(), nil
	}

	config := &scim.ServiceProviderConfig{
		DocumentationURI: optional.NewString("www.example.com/scim"),
	}

	userSchema := schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "userName",
				Required:   true,
				Uniqueness: schema.AttributeUniquenessServer(),
			})),
		},
	}

	var userResourceHandler scim.ResourceHandler
	// initialize w/ own implementation
	resourceTypes := []scim.ResourceType{
		{
			ID:               optional.NewString("User"),
			Name:             "User",
			Endpoint:         "/Users",
			Description:      optional.NewString("User Account"),
			Schema:           userSchema,
			SchemaExtensions: []scim.SchemaExtension{},
			Handler:          userResourceHandler,
		},
	}

	serverArgs := &scim.ServerArgs{
		ServiceProviderConfig: config,
		ResourceTypes:         resourceTypes,
	}

	var serverOpts []scim.ServerOption

	srv, srvErr := scim.NewServer(serverArgs, serverOpts...)
	if srvErr != nil {
		return nil, fmt.Errorf("scim server: %w", srvErr)
	}

	return mw.Handler(srv), nil
}
