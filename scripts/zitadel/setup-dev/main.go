package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/idp"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/project/v2"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func localDevUrl(prefix string) string {
	host := os.Getenv("LOCAL_DEV_HOST")
	if host == "" {
		log.Fatal("LOCAL_DEV_HOST not set")
	}
	return fmt.Sprintf("%s.%s", prefix, host)
}

func main() {
	ctx := context.Background()
	patStr := os.Getenv("ADMIN_SA_PAT")
	if patStr == "" {
		log.Fatal("ADMIN_SA_PAT environment variable empty")
	}
	patAuth := client.PAT(patStr)
	api, clientErr := client.New(ctx, zitadel.New(localDevUrl("auth")), client.WithAuth(patAuth))
	if clientErr != nil {
		log.Fatalf("could not create api client: %s", clientErr)
	}
	if setupErr := setup(ctx, api); setupErr != nil {
		log.Fatalf("setup failed: %s", setupErr)
	}
}

func setup(ctx context.Context, api *client.Client) error {
	orgResp, orgErr := api.ManagementService().GetMyOrg(ctx, &management.GetMyOrgRequest{})
	if orgErr != nil {
		return fmt.Errorf("failed to get org: %w", orgErr)
	}
	org := orgResp.GetOrg()
	orgId := org.GetId()
	log.Printf("Organization is %s (id=%s)", org.GetName(), orgId)

	projectId, projectErr := ensureOrgProject(ctx, api, orgId)
	if projectErr != nil {
		return fmt.Errorf("failed to create project: %w", projectErr)
	}
	log.Printf("Ensured project (id=%s)", projectId)

	idpId, idpErr := ensureTestOIDCProvider(ctx, api)
	if idpErr != nil {
		return fmt.Errorf("failed to ensure test generic oidc provider: %w", idpErr)
	}

	if policyErr := ensureOrgLoginPolicy(ctx, api); policyErr != nil {
		log.Fatalf("ensureOrgLoginPolicy: %v", policyErr)
	}
	log.Printf("Configured login policy: username/password disabled, external IdP enabled")

	if idpPolicyErr := ensureIdpInLoginPolicy(ctx, api, idpId); idpPolicyErr != nil {
		log.Fatalf("ensureIdpInLoginPolicy: %v", idpPolicyErr)
	}

	fmt.Println("DONE")
	fmt.Printf("org_id=%s\nproject_id=%s\norg_oidc_idp_id=%s\n", orgId, projectId, idpId)

	return nil
}

func doIfAlreadyExistsStatus(err error, fn func() error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("failed to get error status: %w", err)
	}
	if st.Code() == codes.AlreadyExists {
		return fn()
	}
	return nil
}

func ensureOrgProject(ctx context.Context, api *client.Client, orgId string) (string, error) {
	name := "rezible"
	projectReq := &project.CreateProjectRequest{
		OrganizationId:         orgId,
		Name:                   name,
		ProjectRoleAssertion:   true,
		AuthorizationRequired:  true,
		ProjectAccessRequired:  true,
		PrivateLabelingSetting: 0,
	}
	projResp, createErr := api.ProjectServiceV2().CreateProject(ctx, projectReq)
	if createErr == nil {
		return projResp.GetProjectId(), nil
	}
	var projectId string
	doErr := doIfAlreadyExistsStatus(createErr, func() error {
		listRes, listErr := api.ProjectServiceV2().ListProjects(ctx, &project.ListProjectsRequest{
			Filters: []*project.ProjectSearchFilter{
				{Filter: &project.ProjectSearchFilter_ProjectNameFilter{
					ProjectNameFilter: &project.ProjectNameFilter{ProjectName: name},
				}},
				{Filter: &project.ProjectSearchFilter_OrganizationIdFilter{
					OrganizationIdFilter: &project.ProjectOrganizationIDFilter{OrganizationId: orgId},
				}},
			},
		})
		if listErr != nil {
			return fmt.Errorf("failed to list projects: %w", listErr)
		}
		if len(listRes.Projects) != 1 {
			return fmt.Errorf("unexpected amount of projects: %v", listRes.Projects)
		}
		projectId = listRes.Projects[0].ProjectId
		return listErr
	})
	return projectId, doErr
}

func ensureTestOIDCProvider(ctx context.Context, api *client.Client) (string, error) {
	name := "Test OIDC"
	provs, listErr := api.ManagementService().ListProviders(ctx, &management.ListProvidersRequest{
		Queries: []*management.ProviderQuery{
			{Query: &management.ProviderQuery_IdpNameQuery{IdpNameQuery: &idp.IDPNameQuery{
				Name: name,
			}}},
		},
	})
	if listErr != nil {
		return "", fmt.Errorf("failed to list providers: %w", listErr)
	}
	if len(provs.Result) >= 1 {
		log.Printf("found %d existing idp providers\n", len(provs.Result))
		return provs.Result[0].GetId(), nil
	}
	addIdpReq := &management.AddGenericOIDCProviderRequest{
		Name:             name,
		Issuer:           "https://" + localDevUrl("oidc-test") + "/",
		ClientId:         os.Getenv("TEST_OIDC_CLIENT_ID"),
		ClientSecret:     os.Getenv("TEST_OIDC_CLIENT_SECRET"),
		Scopes:           []string{"openid", "profile", "email"},
		IsIdTokenMapping: true,
		UsePkce:          true,
	}
	addIdpResp, addIdpErr := api.ManagementService().AddGenericOIDCProvider(ctx, addIdpReq)
	if addIdpErr != nil {
		return "", fmt.Errorf("failed to add: %w", addIdpErr)
	}
	log.Printf("Created org test OIDC IdP: %s", addIdpResp.GetId())
	return addIdpResp.GetId(), nil
}

func ensureOrgLoginPolicy(ctx context.Context, api *client.Client) error {
	req := &management.AddCustomLoginPolicyRequest{
		AllowUsernamePassword:  true,
		AllowRegister:          true,
		AllowExternalIdp:       true,
		ForceMfa:               false,
		HidePasswordReset:      true,
		IgnoreUnknownUsernames: false,
	}
	_, addErr := api.ManagementService().AddCustomLoginPolicy(ctx, req)
	if addErr == nil {
		return nil
	}
	return doIfAlreadyExistsStatus(addErr, func() error {
		//_, upErr := api.ManagementService().UpdateCustomLoginPolicy(ctx, &management.UpdateCustomLoginPolicyRequest{
		//	AllowUsernamePassword:  req.AllowUsernamePassword,
		//	AllowRegister:          req.AllowRegister,
		//	AllowExternalIdp:       req.AllowExternalIdp,
		//	ForceMfa:               req.ForceMfa,
		//	PasswordlessType:       req.PasswordlessType,
		//	HidePasswordReset:      req.HidePasswordReset,
		//	IgnoreUnknownUsernames: req.IgnoreUnknownUsernames,
		//	DefaultRedirectUri:     req.DefaultRedirectUri,
		//	PasswordCheckLifetime:  req.PasswordCheckLifetime,
		//})
		//return upErr
		log.Println("skipping update for custom login policy")
		return nil
	})
}

func ensureIdpInLoginPolicy(ctx context.Context, api *client.Client, idpId string) error {
	pol, polErr := api.ManagementService().GetLoginPolicy(ctx, &management.GetLoginPolicyRequest{})
	if polErr != nil {
		return fmt.Errorf("failed to get login policy: %w", polErr)
	}
	for _, p := range pol.GetPolicy().Idps {
		if p.IdpId == idpId {
			log.Println("idp already in policy")
			return nil
		}
	}

	_, loginIdpErr := api.ManagementService().AddIDPToLoginPolicy(ctx, &management.AddIDPToLoginPolicyRequest{
		IdpId:     idpId,
		OwnerType: idp.IDPOwnerType_IDP_OWNER_TYPE_ORG,
	})
	if loginIdpErr != nil {
		return fmt.Errorf("AddIDPToLoginPolicy: %v", loginIdpErr)
	}
	log.Printf("Attached IdP to org login policy")
	return nil
}
