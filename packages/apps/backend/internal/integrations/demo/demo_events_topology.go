package demoprovider

import (
	"fmt"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/projections"
)

type topologyComponentObservedPayload struct {
	ResourceRef string         `json:"resource_ref"`
	Category    kne.Category   `json:"category"`
	Kind        string         `json:"kind"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
}

func (p topologyComponentObservedPayload) resourceRef() string {
	return p.ResourceRef
}

func (p topologyComponentObservedPayload) getAttributes() projections.SystemComponentEventAttributes {
	return projections.SystemComponentEventAttributes{
		Category:    p.Category,
		Kind:        p.Kind,
		DisplayName: p.DisplayName,
		Description: p.Description,
		Properties:  p.Properties,
	}
}

func componentRef(id string) string {
	return "demo:component:" + id
}

func makeDemoTopologyComponents() []topologyComponentObservedPayload {
	componentAtRef := func(resourceRef string, category kne.Category, kind string, displayName string, description string, properties map[string]any) topologyComponentObservedPayload {
		props := map[string]any{}
		for k, v := range properties {
			props[k] = v
		}
		return topologyComponentObservedPayload{
			ResourceRef: resourceRef,
			Category:    category,
			Kind:        kind,
			DisplayName: displayName,
			Description: description,
			Properties:  props,
		}
	}
	component := func(id string, category kne.Category, kind string, displayName string, description string, properties map[string]any) topologyComponentObservedPayload {
		return componentAtRef(componentRef(id), category, kind, displayName, description, properties)
	}

	return []topologyComponentObservedPayload{
		component("commerce_function", kne.CategorySystemFunction, "capability", "Online Commerce", "Customer-facing commerce capabilities from discovery through payment.", map[string]any{"business_domain": "commerce", "tags": []string{"landscape", "customer-facing"}}),
		component("platform_function", kne.CategorySystemFunction, "capability", "Platform Operations", "Shared platform capabilities that keep commerce systems running.", map[string]any{"business_domain": "platform", "tags": []string{"landscape", "shared"}}),
		component("customer_experience_system", kne.CategorySystem, "subsystem", "Customer Experience", "Web and API entry points for customer and operator journeys.", map[string]any{"tier": "edge", "business_domain": "commerce", "owner_team": "platform_team"}),
		component("commerce_core_system", kne.CategorySystem, "subsystem", "Commerce Core", "Core identity, catalogue, checkout, order, payment, and inventory services.", map[string]any{"tier": "core", "business_domain": "commerce", "owner_team": "commerce_team"}),
		component("platform_services_system", kne.CategorySystem, "subsystem", "Platform Services", "Shared asynchronous, messaging, and operational platform services.", map[string]any{"tier": "supporting", "business_domain": "platform", "owner_team": "platform_team"}),
		component("web_app", kne.CategoryContainer, "user_surface", "Customer Web App", "Primary customer-facing storefront and account experience.", map[string]any{"tier": "edge", "criticality": "high", "lifecycle": "production", "runtime": "sveltekit", "region": "global", "owner_team": "commerce_team", "repository_ref": "rezible-commerce/web-app", "tags": []string{"customer-facing", "frontend"}, "business_domain": "commerce"}),
		component("admin_console", kne.CategoryContainer, "user_surface", "Admin Console", "Internal operations interface for catalog, search, and order support.", map[string]any{"tier": "internal", "criticality": "medium", "lifecycle": "production", "runtime": "sveltekit", "region": "global", "owner_team": "platform_team", "repository_ref": "rezible-commerce/admin-console", "tags": []string{"internal", "operations"}, "business_domain": "operations"}),
		component("public_api_gateway", kne.CategoryContainer, "gateway", "Public API Gateway", "Ingress gateway for public REST and partner API traffic.", map[string]any{"tier": "edge", "criticality": "high", "lifecycle": "production", "runtime": "envoy", "region": "us-east-1", "owner_team": "platform_team", "tags": []string{"api", "ingress"}, "business_domain": "platform"}),
		component("auth_service", kne.CategoryContainer, "service", "Auth Listener", "Authentication, sessions, and customer identity service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "identity_team", "repository_ref": "rezible-commerce/auth-service", "tags": []string{"identity", "sessions"}, "business_domain": "identity"}),
		component("catalog_service", kne.CategoryContainer, "service", "Catalog Listener", "Product catalog read and write API.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_ref": "rezible-commerce/catalog-service", "tags": []string{"products"}, "business_domain": "catalog"}),
		component("search_api", kne.CategoryContainer, "service", "Search API", "Product search query API used by storefront and checkout enrichment.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_ref": "rezible-commerce/search-api", "tags": []string{"search", "customer-facing"}, "business_domain": "catalog"}),
		component("checkout_service", kne.CategoryContainer, "service", "Checkout Listener", "Cart checkout orchestration and payment initiation.", map[string]any{"tier": "core", "criticality": "critical", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_ref": "rezible-commerce/checkout-service", "tags": []string{"checkout", "revenue"}, "business_domain": "checkout"}),
		component("orders_service", kne.CategoryContainer, "service", "Orders Listener", "Order lifecycle and order history service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_ref": "rezible-commerce/orders-service", "tags": []string{"orders"}, "business_domain": "orders"}),
		component("payments_service", kne.CategoryContainer, "service", "Payments Listener", "Payment capture, refunds, and ledger Integration.", map[string]any{"tier": "core", "criticality": "critical", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_ref": "rezible-commerce/payments-service", "tags": []string{"payments", "pci"}, "business_domain": "payments"}),
		component("inventory_service", kne.CategoryContainer, "service", "Inventory Listener", "Stock availability and reservation service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_ref": "rezible-commerce/inventory-service", "tags": []string{"inventory"}, "business_domain": "fulfillment"}),
		component("notifications_service", kne.CategoryContainer, "service", "Notifications Listener", "Customer email and transactional notification API.", map[string]any{"tier": "supporting", "criticality": "medium", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "platform_team", "repository_ref": "rezible-commerce/notifications-service", "tags": []string{"email"}, "business_domain": "communications"}),
		component("search_indexer", kne.CategoryContainer, "worker", "Search Indexer", "Builds and refreshes catalog search indexes.", map[string]any{"tier": "async", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_ref": "rezible-commerce/search-indexer", "tags": []string{"search", "batch"}, "business_domain": "catalog"}),
		component("order_fulfillment_worker", kne.CategoryContainer, "worker", "Order Fulfillment Worker", "Consumes order events and coordinates fulfillment handoff.", map[string]any{"tier": "async", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_ref": "rezible-commerce/order-fulfillment-worker", "tags": []string{"orders", "fulfillment"}, "business_domain": "fulfillment"}),
		component("email_dispatch_worker", kne.CategoryContainer, "worker", "Email Dispatch Worker", "Sends queued transactional customer email.", map[string]any{"tier": "async", "criticality": "medium", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "platform_team", "repository_ref": "rezible-commerce/email-dispatch-worker", "tags": []string{"email", "async"}, "business_domain": "communications"}),
		component("users_postgres", kne.CategoryContainer, "database", "Users Postgres", "Primary user identity database.", map[string]any{"tier": "data", "criticality": "high", "engine": "postgres", "region": "us-east-1", "owner_team": "identity_team", "tags": []string{"identity", "postgres"}}),
		component("catalog_postgres", kne.CategoryContainer, "database", "Catalog Postgres", "System of record for product catalog data.", map[string]any{"tier": "data", "criticality": "high", "engine": "postgres", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"catalog", "postgres"}}),
		component("orders_postgres", kne.CategoryContainer, "database", "Orders Postgres", "System of record for carts, orders, and invoices.", map[string]any{"tier": "data", "criticality": "critical", "engine": "postgres", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"orders", "postgres"}}),
		component("payments_postgres", kne.CategoryContainer, "database", "Payments Postgres", "Payment transaction and ledger database.", map[string]any{"tier": "data", "criticality": "critical", "engine": "postgres", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"payments", "postgres", "pci"}}),
		component("redis_sessions", kne.CategoryContainer, "cache", "Redis Sessions", "Session and short-lived identity cache.", map[string]any{"tier": "data", "criticality": "high", "engine": "redis", "region": "us-east-1", "owner_team": "identity_team", "tags": []string{"cache", "sessions"}}),
		component("redis_search_cache", kne.CategoryContainer, "cache", "Redis Search Cache", "Hot search result and autocomplete cache.", map[string]any{"tier": "data", "criticality": "medium", "engine": "redis", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"cache", "search"}}),
		component("elasticsearch_catalog", kne.CategoryContainer, "search_cluster", "Elasticsearch Catalog", "Primary product search cluster.", map[string]any{"tier": "data", "criticality": "high", "engine": "elasticsearch", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"search", "index"}}),
		component("s3_invoice_bucket", kne.CategoryContainer, "object_store", "Invoice S3 Bucket", "Generated invoice PDF storage.", map[string]any{"tier": "data", "criticality": "medium", "engine": "s3", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"invoices", "documents"}}),
		component("order_events_queue", kne.CategoryContainer, "message_queue", "Order Events Queue", "Durable stream for order lifecycle events.", map[string]any{"tier": "async", "criticality": "high", "engine": "sqs", "region": "us-east-1", "owner_team": "platform_team", "tags": []string{"events", "orders"}}),
		component("stripe", kne.CategorySystem, "external_system", "Stripe", "External payment processor.", map[string]any{"tier": "external", "criticality": "critical", "owner_team": "commerce_team", "tags": []string{"payments", "third-party"}}),
		component("sendgrid", kne.CategorySystem, "external_system", "SendGrid", "External email delivery provider.", map[string]any{"tier": "external", "criticality": "medium", "owner_team": "platform_team", "tags": []string{"email", "third-party"}}),
		component("production_cluster", kne.CategoryInfrastructure, "kubernetes_cluster", "Production Kubernetes", "Shared production cluster hosting the commerce runtime.", map[string]any{"tier": "platform", "criticality": "critical", "provider": "kubernetes", "region": "us-east-1", "owner_team": "platform_team", "tags": []string{"runtime", "shared"}}),
		component("edge_network", kne.CategoryInfrastructure, "load_balancer", "Edge Network", "Shared ingress and load-balancing layer for public and operator traffic.", map[string]any{"tier": "edge", "criticality": "high", "provider": "cloud", "region": "global", "owner_team": "platform_team", "tags": []string{"network", "shared"}}),
		component("shared_observability", kne.CategoryInfrastructure, "telemetry_platform", "Observability Platform", "Shared telemetry and alerting platform for runtime services.", map[string]any{"tier": "platform", "criticality": "high", "provider": "cloud", "region": "global", "owner_team": "platform_team", "tags": []string{"telemetry", "shared"}}),
		component("auth_session_manager", kne.CategoryComponent, "module", "Auth Session Manager", "Internal component that manages sessions and identity state.", map[string]any{"container_ref": "auth_service", "repository_ref": "rezible-commerce/auth-service"}),
		component("search_query_handler", kne.CategoryComponent, "module", "Search Query Handler", "Internal component that serves product and catalogue queries.", map[string]any{"container_ref": "search_api", "repository_ref": "rezible-commerce/search-api"}),
		component("checkout_orchestrator", kne.CategoryComponent, "module", "Checkout Orchestrator", "Internal component coordinating checkout, inventory, payment, and order calls.", map[string]any{"container_ref": "checkout_service", "repository_ref": "rezible-commerce/checkout-service"}),
		component("payment_adapter", kne.CategoryComponent, "module", "Payment Adapter", "Internal component translating checkout requests to the payment provider.", map[string]any{"container_ref": "checkout_service", "repository_ref": "rezible-commerce/checkout-service"}),
		component("catalog_index_job", kne.CategoryComponent, "module", "Catalog Index Job", "Internal component that turns catalogue updates into search index writes.", map[string]any{"container_ref": "search_indexer", "repository_ref": "rezible-commerce/search-indexer"}),
		componentAtRef("demo:code_repositories:search-api", kne.CategoryCode, "repository", "Search API Repository", "Source repository for the search API runtime and query handler.", map[string]any{"repository_ref": "rezible-commerce/search-api", "url": "https://github.example/rezible-commerce/search-api"}),
		componentAtRef("demo:code_repositories:checkout-service", kne.CategoryCode, "repository", "Checkout Service Repository", "Source repository for checkout orchestration and payment integration.", map[string]any{"repository_ref": "rezible-commerce/checkout-service", "url": "https://github.example/rezible-commerce/checkout-service"}),
		componentAtRef("demo:code:checkout-orchestrator", kne.CategoryCode, "source_file", "Checkout Orchestrator Source", "Representative source artifact for checkout orchestration.", map[string]any{"repository_ref": "rezible-commerce/checkout-service", "path": "internal/checkout/orchestrator.go"}),
		componentAtRef("demo:code:search-query-handler", kne.CategoryCode, "source_file", "Search Query Handler Source", "Representative source artifact for search query handling.", map[string]any{"repository_ref": "rezible-commerce/search-api", "path": "internal/search/query_handler.go"}),
		component("identity_team", kne.CategoryActor, "team", "Identity Team", "Owns authentication and customer identity.", map[string]any{"slack_channel": "#team-identity", "oncall_roster": "identity-primary"}),
		component("commerce_team", kne.CategoryActor, "team", "Commerce Team", "Owns catalog, checkout, orders, and payments.", map[string]any{"slack_channel": "#team-commerce", "oncall_roster": "commerce-primary"}),
		component("platform_team", kne.CategoryActor, "team", "Platform Team", "Owns shared platform, messaging, and communications infrastructure.", map[string]any{"slack_channel": "#team-platform", "oncall_roster": "platform-primary"}),
	}
}

type topologyRelationshipObservedPayload struct {
	ResourceRef string                                       `json:"resource_ref"`
	Predicate   knr.Predicate                                `json:"predicate"`
	DisplayName string                                       `json:"display_name,omitempty"`
	Description string                                       `json:"description,omitempty"`
	Properties  map[string]any                               `json:"properties,omitempty"`
	Source      topologyRelationshipObservedPayloadComponent `json:"source"`
	Target      topologyRelationshipObservedPayloadComponent `json:"target"`
}

type topologyRelationshipObservedPayloadComponent struct {
	ResourceRef string       `json:"resource_ref"`
	Category    kne.Category `json:"category"`
	Kind        string       `json:"kind"`
	DisplayName string       `json:"display_name"`
}

func (p topologyRelationshipObservedPayload) resourceRef() string {
	return p.ResourceRef
}

func (p topologyRelationshipObservedPayload) getAttributes(namespace string) projections.SystemRelationshipEventAttributes {
	sourceRef := rez.ProviderResourceRef{
		Provider:          providerName,
		ProviderNamespace: namespace,
		ResourceRef:       p.Source.ResourceRef,
	}
	targetRef := rez.ProviderResourceRef{
		Provider:          providerName,
		ProviderNamespace: namespace,
		ResourceRef:       p.Target.ResourceRef,
	}
	return projections.SystemRelationshipEventAttributes{
		Predicate:   p.Predicate,
		DisplayName: p.DisplayName,
		Description: p.Description,
		Source: projections.EntityObservation{
			Ref:         sourceRef,
			Category:    p.Source.Category,
			Kind:        p.Source.Kind,
			DisplayName: p.Source.DisplayName,
		},
		Target: projections.EntityObservation{
			Ref:         targetRef,
			Category:    p.Target.Category,
			Kind:        p.Target.Kind,
			DisplayName: p.Target.DisplayName,
		},
		Properties: p.Properties,
	}
}

func makeDemoTopologyRelationships(cmps []topologyComponentObservedPayload) []topologyRelationshipObservedPayload {
	resourceRefForID := map[string]string{
		"search_repository":         "demo:code_repositories:search-api",
		"checkout_repository":       "demo:code_repositories:checkout-service",
		"checkout_orchestrator_source": "demo:code:checkout-orchestrator",
		"search_query_source":           "demo:code:search-query-handler",
	}
	mustTopologyComponent := func(id string) topologyRelationshipObservedPayloadComponent {
		ref := componentRef(id)
		if customRef, ok := resourceRefForID[id]; ok {
			ref = customRef
		}
		for _, c := range cmps {
			if c.ResourceRef == ref {
				return topologyRelationshipObservedPayloadComponent{
					ResourceRef: c.ResourceRef,
					Category:    c.Category,
					Kind:        c.Kind,
					DisplayName: c.DisplayName,
				}
			}
		}
		panic(fmt.Sprintf("unknown demo topology component: %s", id))
	}

	rel := func(sourceID string, predicate knr.Predicate, targetID string, displayName string) topologyRelationshipObservedPayload {
		source := mustTopologyComponent(sourceID)
		target := mustTopologyComponent(targetID)
		resourceRef := fmt.Sprintf("demo:relationship:%s:%s:%s", sourceID, predicate, targetID)
		return topologyRelationshipObservedPayload{
			ResourceRef: resourceRef,
			Predicate:   predicate,
			DisplayName: displayName,
			Source:      source,
			Target:      target,
			Properties: map[string]any{
				"resource_ref": resourceRef,
				"source":       sourceID,
				"target":       targetID,
			},
		}
	}

	return []topologyRelationshipObservedPayload{
		rel("commerce_function", knr.PredicateContains, "customer_experience_system", "Online Commerce contains Customer Experience"),
		rel("commerce_function", knr.PredicateContains, "commerce_core_system", "Online Commerce contains Commerce Core"),
		rel("platform_function", knr.PredicateContains, "platform_services_system", "Platform Operations contains Platform Services"),
		rel("customer_experience_system", knr.PredicateContains, "web_app", "Customer Experience contains Customer Web App"),
		rel("customer_experience_system", knr.PredicateContains, "admin_console", "Customer Experience contains Admin Console"),
		rel("customer_experience_system", knr.PredicateContains, "public_api_gateway", "Customer Experience contains Public API Gateway"),
		rel("customer_experience_system", knr.PredicateContains, "search_api", "Customer Experience contains Search API"),
		rel("customer_experience_system", knr.PredicateContains, "edge_network", "Customer Experience contains Edge Network"),
		rel("commerce_core_system", knr.PredicateContains, "auth_service", "Commerce Core contains Auth Listener"),
		rel("commerce_core_system", knr.PredicateContains, "catalog_service", "Commerce Core contains Catalog Listener"),
		rel("commerce_core_system", knr.PredicateContains, "search_api", "Commerce Core contains Search API"),
		rel("commerce_core_system", knr.PredicateContains, "checkout_service", "Commerce Core contains Checkout Listener"),
		rel("commerce_core_system", knr.PredicateContains, "orders_service", "Commerce Core contains Orders Listener"),
		rel("commerce_core_system", knr.PredicateContains, "payments_service", "Commerce Core contains Payments Listener"),
		rel("commerce_core_system", knr.PredicateContains, "inventory_service", "Commerce Core contains Inventory Listener"),
		rel("commerce_core_system", knr.PredicateContains, "users_postgres", "Commerce Core contains Users Postgres"),
		rel("commerce_core_system", knr.PredicateContains, "catalog_postgres", "Commerce Core contains Catalog Postgres"),
		rel("commerce_core_system", knr.PredicateContains, "orders_postgres", "Commerce Core contains Orders Postgres"),
		rel("commerce_core_system", knr.PredicateContains, "payments_postgres", "Commerce Core contains Payments Postgres"),
		rel("commerce_core_system", knr.PredicateContains, "redis_sessions", "Commerce Core contains Redis Sessions"),
		rel("commerce_core_system", knr.PredicateContains, "redis_search_cache", "Commerce Core contains Redis Search Cache"),
		rel("commerce_core_system", knr.PredicateContains, "elasticsearch_catalog", "Commerce Core contains Elasticsearch Catalog"),
		rel("commerce_core_system", knr.PredicateContains, "s3_invoice_bucket", "Commerce Core contains Invoice S3 Bucket"),
		rel("commerce_core_system", knr.PredicateContains, "order_events_queue", "Commerce Core contains Order Events Queue"),
		rel("commerce_core_system", knr.PredicateContains, "production_cluster", "Commerce Core contains Production Kubernetes"),
		rel("platform_services_system", knr.PredicateContains, "notifications_service", "Platform Services contains Notifications Listener"),
		rel("platform_services_system", knr.PredicateContains, "search_indexer", "Platform Services contains Search Indexer"),
		rel("platform_services_system", knr.PredicateContains, "order_fulfillment_worker", "Platform Services contains Order Fulfillment Worker"),
		rel("platform_services_system", knr.PredicateContains, "email_dispatch_worker", "Platform Services contains Email Dispatch Worker"),
		rel("platform_services_system", knr.PredicateContains, "order_events_queue", "Platform Services contains Order Events Queue"),
		rel("platform_services_system", knr.PredicateContains, "production_cluster", "Platform Services contains Production Kubernetes"),
		rel("platform_services_system", knr.PredicateContains, "shared_observability", "Platform Services contains Observability Platform"),
		rel("auth_service", knr.PredicateContains, "auth_session_manager", "Auth Listener contains Auth Session Manager"),
		rel("search_api", knr.PredicateContains, "search_query_handler", "Search API contains Search Query Handler"),
		rel("search_api", knr.PredicateContains, "search_repository", "Search API contains Search API Repository"),
		rel("checkout_service", knr.PredicateContains, "checkout_orchestrator", "Checkout Listener contains Checkout Orchestrator"),
		rel("checkout_service", knr.PredicateContains, "payment_adapter", "Checkout Listener contains Payment Adapter"),
		rel("checkout_service", knr.PredicateContains, "checkout_repository", "Checkout Listener contains Checkout Service Repository"),
		rel("checkout_orchestrator", knr.PredicateContains, "checkout_orchestrator_source", "Checkout Orchestrator contains source"),
		rel("search_query_handler", knr.PredicateContains, "search_query_source", "Search Query Handler contains source"),
		rel("search_indexer", knr.PredicateContains, "catalog_index_job", "Search Indexer contains Catalog Index Job"),
		rel("search_api", knr.PredicateRunsOn, "production_cluster", "Search API runs on Production Kubernetes"),
		rel("checkout_service", knr.PredicateRunsOn, "production_cluster", "Checkout Listener runs on Production Kubernetes"),
		rel("search_indexer", knr.PredicateRunsOn, "production_cluster", "Search Indexer runs on Production Kubernetes"),
		rel("notifications_service", knr.PredicateRunsOn, "production_cluster", "Notifications Listener runs on Production Kubernetes"),
		rel("shared_observability", knr.PredicateObserves, "search_api", "Observability Platform observes Search API"),
		rel("web_app", knr.PredicateCalls, "public_api_gateway", "Customer Web App calls Public API Gateway"),
		rel("admin_console", knr.PredicateCalls, "public_api_gateway", "Admin Console calls Public API Gateway"),
		rel("public_api_gateway", knr.PredicateCalls, "auth_service", "Public API Gateway calls Auth Listener"),
		rel("public_api_gateway", knr.PredicateCalls, "catalog_service", "Public API Gateway calls Catalog Listener"),
		rel("public_api_gateway", knr.PredicateCalls, "search_api", "Public API Gateway calls Search API"),
		rel("public_api_gateway", knr.PredicateCalls, "checkout_service", "Public API Gateway calls Checkout Listener"),
		rel("checkout_service", knr.PredicateCalls, "auth_service", "Checkout Listener calls Auth Listener"),
		rel("checkout_service", knr.PredicateCalls, "search_api", "Checkout Listener calls Search API"),
		rel("checkout_service", knr.PredicateCalls, "inventory_service", "Checkout Listener calls Inventory Listener"),
		rel("checkout_service", knr.PredicateCalls, "payments_service", "Checkout Listener calls Payments Listener"),
		rel("checkout_service", knr.PredicateCalls, "orders_service", "Checkout Listener calls Orders Listener"),
		rel("orders_service", knr.PredicateCalls, "notifications_service", "Orders Listener calls Notifications Listener"),
		rel("auth_service", knr.PredicateReadsFrom, "redis_sessions", "Auth Listener reads from Redis Sessions"),
		rel("auth_service", knr.PredicateWritesTo, "users_postgres", "Auth Listener writes to Users Postgres"),
		rel("catalog_service", knr.PredicateWritesTo, "catalog_postgres", "Catalog Listener writes to Catalog Postgres"),
		rel("catalog_service", knr.PredicatePublishesTo, "order_events_queue", "Catalog Listener publishes to Order Events Queue"),
		rel("search_api", knr.PredicateReadsFrom, "elasticsearch_catalog", "Search API reads from Elasticsearch Catalog"),
		rel("search_api", knr.PredicateReadsFrom, "redis_search_cache", "Search API reads from Redis Search Cache"),
		rel("search_indexer", knr.PredicateReadsFrom, "catalog_postgres", "Search Indexer reads from Catalog Postgres"),
		rel("search_indexer", knr.PredicateWritesTo, "elasticsearch_catalog", "Search Indexer writes to Elasticsearch Catalog"),
		rel("search_indexer", knr.PredicateWritesTo, "redis_search_cache", "Search Indexer writes to Redis Search Cache"),
		rel("orders_service", knr.PredicateWritesTo, "orders_postgres", "Orders Listener writes to Orders Postgres"),
		rel("orders_service", knr.PredicatePublishesTo, "order_events_queue", "Orders Listener publishes to Order Events Queue"),
		rel("payments_service", knr.PredicateWritesTo, "payments_postgres", "Payments Listener writes to Payments Postgres"),
		rel("payments_service", knr.PredicateCalls, "stripe", "Payments Listener calls Stripe"),
		rel("inventory_service", knr.PredicateReadsFrom, "catalog_postgres", "Inventory Listener reads from Catalog Postgres"),
		rel("order_fulfillment_worker", knr.PredicateConsumesFrom, "order_events_queue", "Order Fulfillment Worker consumes from Order Events Queue"),
		rel("email_dispatch_worker", knr.PredicateConsumesFrom, "order_events_queue", "Email Dispatch Worker consumes from Order Events Queue"),
		rel("notifications_service", knr.PredicatePublishesTo, "order_events_queue", "Notifications Listener publishes to Order Events Queue"),
		rel("notifications_service", knr.PredicateCalls, "sendgrid", "Notifications Listener calls SendGrid"),
		rel("orders_service", knr.PredicateWritesTo, "s3_invoice_bucket", "Orders Listener writes to Invoice S3 Bucket"),
		rel("identity_team", knr.PredicateOwns, "auth_service", "Identity Team owns Auth Listener"),
		rel("identity_team", knr.PredicateOwns, "users_postgres", "Identity Team owns Users Postgres"),
		rel("commerce_team", knr.PredicateOwns, "catalog_service", "Commerce Team owns Catalog Listener"),
		rel("commerce_team", knr.PredicateOwns, "search_api", "Commerce Team owns Search API"),
		rel("commerce_team", knr.PredicateOwns, "checkout_service", "Commerce Team owns Checkout Listener"),
		rel("commerce_team", knr.PredicateOwns, "orders_service", "Commerce Team owns Orders Listener"),
		rel("commerce_team", knr.PredicateOwns, "payments_service", "Commerce Team owns Payments Listener"),
		rel("platform_team", knr.PredicateOwns, "public_api_gateway", "Platform Team owns Public API Gateway"),
		rel("platform_team", knr.PredicateOwns, "notifications_service", "Platform Team owns Notifications Listener"),
	}
}
