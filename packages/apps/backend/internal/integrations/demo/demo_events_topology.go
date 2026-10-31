package demoprovider

import (
	"fmt"
	"slices"

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
		component("customer_experience_system", kne.CategorySystem, "subsystem", "Customer Experience", "Web and API entry points for customer and operator journeys.", map[string]any{"tier": "edge", "business_domain": "commerce"}),
		component("commerce_core_system", kne.CategorySystem, "subsystem", "Commerce Core", "Core identity, catalogue, checkout, order, payment, and inventory services.", map[string]any{"tier": "core", "business_domain": "commerce"}),
		component("platform_services_system", kne.CategorySystem, "subsystem", "Platform Services", "Shared asynchronous, messaging, and operational platform services.", map[string]any{"tier": "supporting", "business_domain": "platform"}),
		component("web_app", kne.CategoryContainer, "web_app", "Customer Web App", "Primary customer-facing storefront and account experience.", map[string]any{"tier": "edge", "criticality": "high", "lifecycle": "production", "runtime": "sveltekit", "region": "global", "tags": []string{"customer-facing", "frontend"}, "business_domain": "commerce"}),
		component("admin_console", kne.CategoryContainer, "web_app", "Admin Console", "Internal operations interface for catalog, search, and order support.", map[string]any{"tier": "internal", "criticality": "medium", "lifecycle": "production", "runtime": "sveltekit", "region": "global", "tags": []string{"internal", "operations"}, "business_domain": "operations"}),
		component("public_api_gateway", kne.CategoryContainer, "gateway", "Public API Gateway", "Ingress gateway for public REST and partner API traffic.", map[string]any{"tier": "edge", "criticality": "high", "lifecycle": "production", "runtime": "envoy", "region": "us-east-1", "tags": []string{"api", "ingress"}, "business_domain": "platform"}),
		component("auth_service", kne.CategoryContainer, "service", "Auth Listener", "Authentication, sessions, and customer identity service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"identity", "sessions"}, "business_domain": "identity"}),
		component("catalog_service", kne.CategoryContainer, "service", "Catalog Listener", "Product catalog read and write API.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"products"}, "business_domain": "catalog"}),
		component("search_api", kne.CategoryContainer, "service", "Search API", "Product search query API used by storefront and checkout enrichment.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"search", "customer-facing"}, "business_domain": "catalog"}),
		component("checkout_service", kne.CategoryContainer, "service", "Checkout Listener", "Cart checkout orchestration and payment initiation.", map[string]any{"tier": "core", "criticality": "critical", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"checkout", "revenue"}, "business_domain": "checkout"}),
		component("orders_service", kne.CategoryContainer, "service", "Orders Listener", "Order lifecycle and order history service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"orders"}, "business_domain": "orders"}),
		component("payments_service", kne.CategoryContainer, "service", "Payments Listener", "Payment capture, refunds, and ledger Integration.", map[string]any{"tier": "core", "criticality": "critical", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"payments", "pci"}, "business_domain": "payments"}),
		component("inventory_service", kne.CategoryContainer, "service", "Inventory Listener", "Stock availability and reservation service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"inventory"}, "business_domain": "fulfillment"}),
		component("notifications_service", kne.CategoryContainer, "service", "Notifications Listener", "Customer email and transactional notification API.", map[string]any{"tier": "supporting", "criticality": "medium", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"email"}, "business_domain": "communications"}),
		component("search_indexer", kne.CategoryContainer, "worker", "Search Indexer", "Builds and refreshes catalog search indexes.", map[string]any{"tier": "async", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"search", "batch"}, "business_domain": "catalog"}),
		component("order_fulfillment_worker", kne.CategoryContainer, "worker", "Order Fulfillment Worker", "Consumes order events and coordinates fulfillment handoff.", map[string]any{"tier": "async", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"orders", "fulfillment"}, "business_domain": "fulfillment"}),
		component("email_dispatch_worker", kne.CategoryContainer, "worker", "Email Dispatch Worker", "Sends queued transactional customer email.", map[string]any{"tier": "async", "criticality": "medium", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "tags": []string{"email", "async"}, "business_domain": "communications"}),
		component("users_postgres", kne.CategoryContainer, "database", "Users Postgres", "Primary user identity database.", map[string]any{"tier": "data", "criticality": "high", "engine": "postgres", "region": "us-east-1", "tags": []string{"identity", "postgres"}}),
		component("catalog_postgres", kne.CategoryContainer, "database", "Catalog Postgres", "System of record for product catalog data.", map[string]any{"tier": "data", "criticality": "high", "engine": "postgres", "region": "us-east-1", "tags": []string{"catalog", "postgres"}}),
		component("orders_postgres", kne.CategoryContainer, "database", "Orders Postgres", "System of record for carts, orders, and invoices.", map[string]any{"tier": "data", "criticality": "critical", "engine": "postgres", "region": "us-east-1", "tags": []string{"orders", "postgres"}}),
		component("payments_postgres", kne.CategoryContainer, "database", "Payments Postgres", "Payment transaction and ledger database.", map[string]any{"tier": "data", "criticality": "critical", "engine": "postgres", "region": "us-east-1", "tags": []string{"payments", "postgres", "pci"}}),
		component("redis_sessions", kne.CategoryContainer, "cache", "Redis Sessions", "Session and short-lived identity cache.", map[string]any{"tier": "data", "criticality": "high", "engine": "redis", "region": "us-east-1", "tags": []string{"cache", "sessions"}}),
		component("redis_search_cache", kne.CategoryContainer, "cache", "Redis Search Cache", "Hot search result and autocomplete cache.", map[string]any{"tier": "data", "criticality": "medium", "engine": "redis", "region": "us-east-1", "tags": []string{"cache", "search"}}),
		component("elasticsearch_catalog", kne.CategoryContainer, "search_cluster", "Elasticsearch Catalog", "Primary product search cluster.", map[string]any{"tier": "data", "criticality": "high", "engine": "elasticsearch", "region": "us-east-1", "tags": []string{"search", "index"}}),
		component("s3_invoice_bucket", kne.CategoryContainer, "object_store", "Invoice S3 Bucket", "Generated invoice PDF storage.", map[string]any{"tier": "data", "criticality": "medium", "engine": "s3", "region": "us-east-1", "tags": []string{"invoices", "documents"}}),
		component("order_events_queue", kne.CategoryContainer, "message_queue", "Order Events Queue", "Durable stream for order lifecycle events.", map[string]any{"tier": "async", "criticality": "high", "engine": "sqs", "region": "us-east-1", "tags": []string{"events", "orders"}}),
		component("stripe", kne.CategorySystem, "external_system", "Stripe", "External payment processor.", map[string]any{"tier": "external", "criticality": "critical", "tags": []string{"payments", "third-party"}}),
		component("sendgrid", kne.CategorySystem, "external_system", "SendGrid", "External email delivery provider.", map[string]any{"tier": "external", "criticality": "medium", "tags": []string{"email", "third-party"}}),
		component("production", kne.CategoryInfrastructure, "environment", "production", "Production environment for every customer-facing workload.", map[string]any{"tier": "platform", "criticality": "critical", "tags": []string{"runtime"}}),
		component("prod_use1", kne.CategoryInfrastructure, "kubernetes_cluster", "prod-use1", "Shared production Kubernetes cluster in us-east-1.", map[string]any{"tier": "platform", "criticality": "critical", "provider": "kubernetes", "region": "us-east-1", "tags": []string{"runtime", "shared"}}),
		component("commerce_namespace", kne.CategoryInfrastructure, "kubernetes_namespace", "commerce", "Namespace for the commerce workloads.", nil),
		component("platform_namespace", kne.CategoryInfrastructure, "kubernetes_namespace", "platform", "Namespace for the shared platform workloads.", nil),
		component("edge_network", kne.CategoryInfrastructure, "load_balancer", "Edge Network", "Shared ingress and load-balancing layer for public and operator traffic.", map[string]any{"tier": "edge", "criticality": "high", "provider": "cloud", "region": "global", "tags": []string{"network", "shared"}}),
		component("auth_session_manager", kne.CategoryComponent, "module", "Auth Session Manager", "Internal component that manages sessions and identity state.", nil),
		component("search_query_handler", kne.CategoryComponent, "module", "Search Query Handler", "Internal component that serves product and catalogue queries.", nil),
		component("checkout_orchestrator", kne.CategoryComponent, "module", "Checkout Orchestrator", "Internal component coordinating checkout, inventory, payment, and order calls.", nil),
		component("payment_adapter", kne.CategoryComponent, "module", "Payment Adapter", "Internal component translating checkout requests to the payment provider.", nil),
		component("catalog_index_job", kne.CategoryComponent, "module", "Catalog Index Job", "Internal component that turns catalogue updates into search index writes.", nil),
		componentAtRef("demo:code_repositories:search-api", kne.CategoryCode, "repository", "Search API Repository", "Source repository for the search API runtime and query handler.", map[string]any{"url": "https://github.example/rezible-commerce/search-api"}),
		componentAtRef("demo:code_repositories:checkout-service", kne.CategoryCode, "repository", "Checkout Service Repository", "Source repository for checkout orchestration and payment integration.", map[string]any{"url": "https://github.example/rezible-commerce/checkout-service"}),
		componentAtRef("demo:code:checkout-orchestrator", kne.CategoryCode, "source_file", "Checkout Orchestrator Source", "Representative source artifact for checkout orchestration.", map[string]any{"path": "internal/checkout/orchestrator.go"}),
		componentAtRef("demo:code:search-query-handler", kne.CategoryCode, "source_file", "Search Query Handler Source", "Representative source artifact for search query handling.", map[string]any{"path": "internal/search/query_handler.go"}),
		component("identity_team", kne.CategoryActor, "team", "Identity Team", "Owns authentication and customer identity.", map[string]any{"slack_channel": "#team-identity", "oncall_roster": "identity-primary"}),
		component("commerce_team", kne.CategoryActor, "team", "Commerce Team", "Owns catalog, checkout, orders, and payments.", map[string]any{"slack_channel": "#team-commerce", "oncall_roster": "commerce-primary"}),
		component("platform_team", kne.CategoryActor, "team", "Platform Team", "Owns shared platform, messaging, and communications infrastructure.", map[string]any{"slack_channel": "#team-platform", "oncall_roster": "platform-primary"}),
		component("engineering", kne.CategoryActor, "department", "Engineering", "The engineering department.", nil),
		component("customers", kne.CategoryActor, "customer_segment", "Customers", "Shoppers using the storefront.", nil),
		component("checkout_flow", kne.CategoryProcess, "business_process", "Checkout flow", "From cart to captured payment.", nil),
		component("payment_capture_step", kne.CategoryProcess, "step", "Payment capture", "Capturing the authorised payment.", nil),
		component("duplicate_payment_capture", kne.CategoryConcern, "risk", "Duplicate payment capture", "A retried capture charges the customer twice.", map[string]any{"severity": "high"}),
		component("idempotent_capture_adr", kne.CategoryDecision, "adr", "Use idempotency keys for capture", "Every capture request carries an idempotency key.", nil),
		component("checkout_dashboard", kne.CategorySignal, "dashboard", "Checkout dashboard", "Checkout request rate, errors and latency.", nil),
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
		Provider:          ProviderName,
		ProviderNamespace: namespace,
		ResourceRef:       p.Source.ResourceRef,
	}
	targetRef := rez.ProviderResourceRef{
		Provider:          ProviderName,
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
		"search_repository":            "demo:code_repositories:search-api",
		"checkout_repository":          "demo:code_repositories:checkout-service",
		"checkout_orchestrator_source": "demo:code:checkout-orchestrator",
		"search_query_source":          "demo:code:search-query-handler",
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

	relWith := func(sourceID string, predicate knr.Predicate, targetID string, displayName string, properties map[string]any) topologyRelationshipObservedPayload {
		resourceRef := fmt.Sprintf("demo:relationship:%s:%s:%s", sourceID, predicate, targetID)
		return topologyRelationshipObservedPayload{
			ResourceRef: resourceRef,
			Predicate:   predicate,
			DisplayName: displayName,
			Source:      mustTopologyComponent(sourceID),
			Target:      mustTopologyComponent(targetID),
			Properties:  properties,
		}
	}
	rel := func(sourceID string, predicate knr.Predicate, targetID string, displayName string) topologyRelationshipObservedPayload {
		return relWith(sourceID, predicate, targetID, displayName, nil)
	}
	contains := func(parentID string, childIDs ...string) []topologyRelationshipObservedPayload {
		rels := make([]topologyRelationshipObservedPayload, 0, len(childIDs))
		for _, childID := range childIDs {
			name := mustTopologyComponent(parentID).DisplayName + " contains " + mustTopologyComponent(childID).DisplayName
			rels = append(rels, rel(parentID, knr.PredicateContains, childID, name))
		}
		return rels
	}
	runsOn := func(targetID string, sourceIDs ...string) []topologyRelationshipObservedPayload {
		rels := make([]topologyRelationshipObservedPayload, 0, len(sourceIDs))
		for _, sourceID := range sourceIDs {
			name := mustTopologyComponent(sourceID).DisplayName + " runs on " + mustTopologyComponent(targetID).DisplayName
			rels = append(rels, rel(sourceID, knr.PredicateRunsOn, targetID, name))
		}
		return rels
	}

	structure := slices.Concat(
		// Logical
		contains("commerce_function", "customer_experience_system", "commerce_core_system"),
		contains("platform_function", "platform_services_system"),
		contains("customer_experience_system", "web_app", "admin_console", "public_api_gateway"),
		contains("commerce_core_system",
			"auth_service", "catalog_service", "search_api", "checkout_service", "orders_service", "payments_service",
			"inventory_service", "users_postgres", "catalog_postgres", "orders_postgres", "payments_postgres",
			"redis_sessions", "redis_search_cache", "elasticsearch_catalog", "s3_invoice_bucket"),
		contains("platform_services_system",
			"notifications_service", "search_indexer", "order_fulfillment_worker", "email_dispatch_worker", "order_events_queue"),
		contains("auth_service", "auth_session_manager"),
		contains("search_api", "search_query_handler"),
		contains("checkout_service", "checkout_orchestrator", "payment_adapter"),
		contains("search_indexer", "catalog_index_job"),
		// Deployment
		contains("production", "prod_use1", "edge_network"),
		contains("prod_use1", "commerce_namespace", "platform_namespace"),
		runsOn("commerce_namespace",
			"web_app", "auth_service", "catalog_service", "search_api", "checkout_service", "orders_service",
			"payments_service", "inventory_service"),
		runsOn("platform_namespace",
			"admin_console", "public_api_gateway", "notifications_service", "search_indexer", "order_fulfillment_worker",
			"email_dispatch_worker"),
		runsOn("production",
			"users_postgres", "catalog_postgres", "orders_postgres", "payments_postgres", "redis_sessions",
			"redis_search_cache", "elasticsearch_catalog", "s3_invoice_bucket", "order_events_queue"),
		// Code
		contains("checkout_repository", "checkout_orchestrator_source"),
		contains("search_repository", "search_query_source"),
		// Organization
		contains("engineering", "identity_team", "commerce_team", "platform_team"),
		// Process
		contains("checkout_flow", "payment_capture_step"),
	)

	return append(structure,
		rel("checkout_service", knr.PredicateBuiltFrom, "checkout_repository", "Checkout Listener is built from Checkout Service Repository"),
		rel("checkout_orchestrator", knr.PredicateBuiltFrom, "checkout_orchestrator_source", "Checkout Orchestrator is built from its source"),
		rel("search_api", knr.PredicateBuiltFrom, "search_repository", "Search API is built from Search API Repository"),
		rel("search_query_handler", knr.PredicateBuiltFrom, "search_query_source", "Search Query Handler is built from its source"),
		rel("checkout_dashboard", knr.PredicateObserves, "checkout_service", "Checkout dashboard observes Checkout Listener"),
		rel("customers", knr.PredicateUses, "web_app", "Customers use Customer Web App"),
		rel("edge_network", knr.PredicateCalls, "public_api_gateway", "Edge Network calls Public API Gateway"),
		rel("web_app", knr.PredicateCalls, "public_api_gateway", "Customer Web App calls Public API Gateway"),
		rel("admin_console", knr.PredicateCalls, "public_api_gateway", "Admin Console calls Public API Gateway"),
		rel("admin_console", knr.PredicateDependsOn, "auth_service", "Admin Console depends on Auth Listener"),
		rel("public_api_gateway", knr.PredicateCalls, "auth_service", "Public API Gateway calls Auth Listener"),
		rel("public_api_gateway", knr.PredicateCalls, "catalog_service", "Public API Gateway calls Catalog Listener"),
		rel("public_api_gateway", knr.PredicateCalls, "search_api", "Public API Gateway calls Search API"),
		rel("public_api_gateway", knr.PredicateCalls, "checkout_service", "Public API Gateway calls Checkout Listener"),
		rel("checkout_service", knr.PredicateCalls, "auth_service", "Checkout Listener calls Auth Listener"),
		relWith("checkout_service", knr.PredicateCalls, "search_api", "Checkout Listener calls Search API", map[string]any{"optional": true, "timeout_ms": 200}),
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
		rel("commerce_team", knr.PredicateParticipatesIn, "checkout_flow", "Commerce Team participates in Checkout flow"),
		rel("web_app", knr.PredicateSupports, "checkout_flow", "Customer Web App supports Checkout flow"),
		rel("checkout_service", knr.PredicateSupports, "checkout_flow", "Checkout Listener supports Checkout flow"),
		rel("payments_service", knr.PredicateSupports, "payment_capture_step", "Payments Listener supports Payment capture"),
		rel("payments_service", knr.PredicateMitigates, "duplicate_payment_capture", "Payments Listener mitigates Duplicate payment capture"),
		rel("idempotent_capture_adr", knr.PredicateAddresses, "duplicate_payment_capture", "Idempotency keys address Duplicate payment capture"),
		rel("idempotent_capture_adr", knr.PredicateConstrains, "payments_service", "Idempotency keys constrain Payments Listener"),
	)
}
