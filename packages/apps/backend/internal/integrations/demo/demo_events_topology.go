package demoprovider

import (
	"fmt"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/projections"
)

type topologyComponentObservedPayload struct {
	ExternalRef string         `json:"external_ref"`
	Category    kne.Category   `json:"category"`
	Kind        string         `json:"kind"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
}

func (p topologyComponentObservedPayload) subjectRef() string {
	return p.ExternalRef
}

func (p topologyComponentObservedPayload) getAttributes() projections.SystemComponentSubjectAttributes {
	return projections.SystemComponentSubjectAttributes{
		ExternalRef: p.ExternalRef,
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
	component := func(id string, category kne.Category, kind string, displayName string, description string, properties map[string]any) topologyComponentObservedPayload {
		props := map[string]any{}
		for k, v := range properties {
			props[k] = v
		}
		return topologyComponentObservedPayload{
			ExternalRef: componentRef(id),
			Category:    category,
			Kind:        kind,
			DisplayName: displayName,
			Description: description,
			Properties:  props,
		}
	}

	return []topologyComponentObservedPayload{
		component("web_app", kne.CategoryContainer, "user_surface", "Customer Web App", "Primary customer-facing storefront and account experience.", map[string]any{"tier": "edge", "criticality": "high", "lifecycle": "production", "runtime": "sveltekit", "region": "global", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/web-app", "tags": []string{"customer-facing", "frontend"}, "business_domain": "commerce"}),
		component("admin_console", kne.CategoryContainer, "user_surface", "Admin Console", "Internal operations interface for catalog, search, and order support.", map[string]any{"tier": "internal", "criticality": "medium", "lifecycle": "production", "runtime": "sveltekit", "region": "global", "owner_team": "platform_team", "repository_external_ref": "rezible-commerce/admin-console", "tags": []string{"internal", "operations"}, "business_domain": "operations"}),
		component("public_api_gateway", kne.CategoryContainer, "gateway", "Public API Gateway", "Ingress gateway for public REST and partner API traffic.", map[string]any{"tier": "edge", "criticality": "high", "lifecycle": "production", "runtime": "envoy", "region": "us-east-1", "owner_team": "platform_team", "tags": []string{"api", "ingress"}, "business_domain": "platform"}),
		component("auth_service", kne.CategoryContainer, "service", "Auth Listener", "Authentication, sessions, and customer identity service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "identity_team", "repository_external_ref": "rezible-commerce/auth-service", "tags": []string{"identity", "sessions"}, "business_domain": "identity"}),
		component("catalog_service", kne.CategoryContainer, "service", "Catalog Listener", "Product catalog read and write API.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/catalog-service", "tags": []string{"products"}, "business_domain": "catalog"}),
		component("search_api", kne.CategoryContainer, "service", "Search API", "Product search query API used by storefront and checkout enrichment.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/search-api", "tags": []string{"search", "customer-facing"}, "business_domain": "catalog"}),
		component("checkout_service", kne.CategoryContainer, "service", "Checkout Listener", "Cart checkout orchestration and payment initiation.", map[string]any{"tier": "core", "criticality": "critical", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/checkout-service", "tags": []string{"checkout", "revenue"}, "business_domain": "checkout"}),
		component("orders_service", kne.CategoryContainer, "service", "Orders Listener", "Order lifecycle and order history service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/orders-service", "tags": []string{"orders"}, "business_domain": "orders"}),
		component("payments_service", kne.CategoryContainer, "service", "Payments Listener", "Payment capture, refunds, and ledger Integration.", map[string]any{"tier": "core", "criticality": "critical", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/payments-service", "tags": []string{"payments", "pci"}, "business_domain": "payments"}),
		component("inventory_service", kne.CategoryContainer, "service", "Inventory Listener", "Stock availability and reservation service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/inventory-service", "tags": []string{"inventory"}, "business_domain": "fulfillment"}),
		component("notifications_service", kne.CategoryContainer, "service", "Notifications Listener", "Customer email and transactional notification API.", map[string]any{"tier": "supporting", "criticality": "medium", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "platform_team", "repository_external_ref": "rezible-commerce/notifications-service", "tags": []string{"email"}, "business_domain": "communications"}),
		component("search_indexer", kne.CategoryContainer, "worker", "Search Indexer", "Builds and refreshes catalog search indexes.", map[string]any{"tier": "async", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/search-indexer", "tags": []string{"search", "batch"}, "business_domain": "catalog"}),
		component("order_fulfillment_worker", kne.CategoryContainer, "worker", "Order Fulfillment Worker", "Consumes order events and coordinates fulfillment handoff.", map[string]any{"tier": "async", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/order-fulfillment-worker", "tags": []string{"orders", "fulfillment"}, "business_domain": "fulfillment"}),
		component("email_dispatch_worker", kne.CategoryContainer, "worker", "Email Dispatch Worker", "Sends queued transactional customer email.", map[string]any{"tier": "async", "criticality": "medium", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "platform_team", "repository_external_ref": "rezible-commerce/email-dispatch-worker", "tags": []string{"email", "async"}, "business_domain": "communications"}),
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
		component("customer", kne.CategoryDomainObject, "business_entity", "Customer", "A person or organization buying from the storefront.", map[string]any{"business_domain": "identity", "owner_team": "identity_team", "tags": []string{"model"}}),
		component("product", kne.CategoryDomainObject, "business_entity", "Product", "A sellable catalog item.", map[string]any{"business_domain": "catalog", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("cart", kne.CategoryDomainObject, "business_entity", "Cart", "A customer's active purchase intent.", map[string]any{"business_domain": "checkout", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("order", kne.CategoryDomainObject, "business_entity", "Order", "A committed customer purchase.", map[string]any{"business_domain": "orders", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("payment", kne.CategoryDomainObject, "business_entity", "Payment", "Payment authorization and capture record.", map[string]any{"business_domain": "payments", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("invoice", kne.CategoryDomainObject, "business_entity", "Invoice", "Customer invoice document.", map[string]any{"business_domain": "orders", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("search_index", kne.CategoryDomainObject, "business_entity", "Search Index", "Materialized product search index.", map[string]any{"business_domain": "catalog", "owner_team": "commerce_team", "tags": []string{"model", "derived"}}),
		component("identity_team", kne.CategoryActor, "team", "Identity Team", "Owns authentication and customer identity.", map[string]any{"slack_channel": "#team-identity", "oncall_roster": "identity-primary"}),
		component("commerce_team", kne.CategoryActor, "team", "Commerce Team", "Owns catalog, checkout, orders, and payments.", map[string]any{"slack_channel": "#team-commerce", "oncall_roster": "commerce-primary"}),
		component("platform_team", kne.CategoryActor, "team", "Platform Team", "Owns shared platform, messaging, and communications infrastructure.", map[string]any{"slack_channel": "#team-platform", "oncall_roster": "platform-primary"}),
	}
}

type topologyRelationshipObservedPayload struct {
	ExternalRef string                                       `json:"external_ref"`
	Predicate   knr.Predicate                                `json:"predicate"`
	DisplayName string                                       `json:"display_name,omitempty"`
	Description string                                       `json:"description,omitempty"`
	Properties  map[string]any                               `json:"properties,omitempty"`
	Source      topologyRelationshipObservedPayloadComponent `json:"source"`
	Target      topologyRelationshipObservedPayloadComponent `json:"target"`
}

type topologyRelationshipObservedPayloadComponent struct {
	ExternalRef string       `json:"external_ref"`
	Category    kne.Category `json:"category"`
	Kind        string       `json:"kind"`
	DisplayName string       `json:"display_name"`
}

func (p topologyRelationshipObservedPayload) subjectRef() string {
	return p.ExternalRef
}

func (p topologyRelationshipObservedPayload) getAttributes() projections.SystemRelationshipSubjectAttributes {
	return projections.SystemRelationshipSubjectAttributes{
		ExternalRef:       p.ExternalRef,
		Predicate:         p.Predicate,
		DisplayName:       p.DisplayName,
		Description:       p.Description,
		SourceExternalRef: p.Source.ExternalRef,
		SourceCategory:    p.Source.Category,
		SourceKind:        p.Source.Kind,
		SourceDisplayName: p.Source.DisplayName,
		TargetExternalRef: p.Target.ExternalRef,
		TargetCategory:    p.Target.Category,
		TargetKind:        p.Target.Kind,
		TargetDisplayName: p.Target.DisplayName,
		Properties:        p.Properties,
	}
}

func makeDemoTopologyRelationships(cmps []topologyComponentObservedPayload) []topologyRelationshipObservedPayload {
	mustTopologyComponent := func(id string) topologyRelationshipObservedPayloadComponent {
		ref := componentRef(id)
		for _, c := range cmps {
			if c.ExternalRef == ref {
				return topologyRelationshipObservedPayloadComponent{
					ExternalRef: c.ExternalRef,
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
		externalRef := fmt.Sprintf("demo:relationship:%s:%s:%s", sourceID, predicate, targetID)
		return topologyRelationshipObservedPayload{
			ExternalRef: externalRef,
			Predicate:   predicate,
			DisplayName: displayName,
			Source:      source,
			Target:      target,
			Properties: map[string]any{
				"external_ref": externalRef,
				"source":       sourceID,
				"target":       targetID,
			},
		}
	}

	return []topologyRelationshipObservedPayload{
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
		rel("auth_service", knr.PredicateProcesses, "customer", "Auth Listener processes Customer"),
		rel("catalog_service", knr.PredicateProcesses, "product", "Catalog Listener processes Product"),
		rel("checkout_service", knr.PredicateProcesses, "cart", "Checkout Listener processes Cart"),
		rel("orders_service", knr.PredicateProcesses, "order", "Orders Listener processes Order"),
		rel("payments_service", knr.PredicateProcesses, "payment", "Payments Listener processes Payment"),
		rel("orders_service", knr.PredicateProcesses, "invoice", "Orders Listener processes Invoice"),
		rel("search_indexer", knr.PredicateIndexes, "product", "Search Indexer indexes Product"),
		rel("elasticsearch_catalog", knr.PredicateStores, "search_index", "Elasticsearch Catalog stores Search Index"),
	}
}
