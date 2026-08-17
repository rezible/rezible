package demoprovider

import (
	"fmt"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/projections"
)

type topologyComponentObservedPayload struct {
	ExternalRef string         `json:"external_ref"`
	Kind        kne.Kind       `json:"kind"`
	Subkind     string         `json:"subkind"`
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
		Kind:        p.Kind,
		Subkind:     p.Subkind,
		DisplayName: p.DisplayName,
		Description: p.Description,
		Properties:  p.Properties,
	}
}

func componentRef(id string) string {
	return "demo:component:" + id
}

func makeDemoTopologyComponents() []topologyComponentObservedPayload {
	component := func(id string, kind kne.Kind, subkind string, displayName string, description string, properties map[string]any) topologyComponentObservedPayload {
		props := map[string]any{}
		for k, v := range properties {
			props[k] = v
		}
		return topologyComponentObservedPayload{
			ExternalRef: componentRef(id),
			Kind:        kind,
			Subkind:     subkind,
			DisplayName: displayName,
			Description: description,
			Properties:  props,
		}
	}

	return []topologyComponentObservedPayload{
		component("web_app", kne.KindContainer, "user_surface", "Customer Web App", "Primary customer-facing storefront and account experience.", map[string]any{"tier": "edge", "criticality": "high", "lifecycle": "production", "runtime": "sveltekit", "region": "global", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/web-app", "tags": []string{"customer-facing", "frontend"}, "business_domain": "commerce"}),
		component("admin_console", kne.KindContainer, "user_surface", "Admin Console", "Internal operations interface for catalog, search, and order support.", map[string]any{"tier": "internal", "criticality": "medium", "lifecycle": "production", "runtime": "sveltekit", "region": "global", "owner_team": "platform_team", "repository_external_ref": "rezible-commerce/admin-console", "tags": []string{"internal", "operations"}, "business_domain": "operations"}),
		component("public_api_gateway", kne.KindContainer, "gateway", "Public API Gateway", "Ingress gateway for public REST and partner API traffic.", map[string]any{"tier": "edge", "criticality": "high", "lifecycle": "production", "runtime": "envoy", "region": "us-east-1", "owner_team": "platform_team", "tags": []string{"api", "ingress"}, "business_domain": "platform"}),
		component("auth_service", kne.KindContainer, "service", "Auth Listener", "Authentication, sessions, and customer identity service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "identity_team", "repository_external_ref": "rezible-commerce/auth-service", "tags": []string{"identity", "sessions"}, "business_domain": "identity"}),
		component("catalog_service", kne.KindContainer, "service", "Catalog Listener", "Product catalog read and write API.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/catalog-service", "tags": []string{"products"}, "business_domain": "catalog"}),
		component("search_api", kne.KindContainer, "service", "Search API", "Product search query API used by storefront and checkout enrichment.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/search-api", "tags": []string{"search", "customer-facing"}, "business_domain": "catalog"}),
		component("checkout_service", kne.KindContainer, "service", "Checkout Listener", "Cart checkout orchestration and payment initiation.", map[string]any{"tier": "core", "criticality": "critical", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/checkout-service", "tags": []string{"checkout", "revenue"}, "business_domain": "checkout"}),
		component("orders_service", kne.KindContainer, "service", "Orders Listener", "Order lifecycle and order history service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/orders-service", "tags": []string{"orders"}, "business_domain": "orders"}),
		component("payments_service", kne.KindContainer, "service", "Payments Listener", "Payment capture, refunds, and ledger Integration.", map[string]any{"tier": "core", "criticality": "critical", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/payments-service", "tags": []string{"payments", "pci"}, "business_domain": "payments"}),
		component("inventory_service", kne.KindContainer, "service", "Inventory Listener", "Stock availability and reservation service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/inventory-service", "tags": []string{"inventory"}, "business_domain": "fulfillment"}),
		component("notifications_service", kne.KindContainer, "service", "Notifications Listener", "Customer email and transactional notification API.", map[string]any{"tier": "supporting", "criticality": "medium", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "platform_team", "repository_external_ref": "rezible-commerce/notifications-service", "tags": []string{"email"}, "business_domain": "communications"}),
		component("search_indexer", kne.KindContainer, "worker", "Search Indexer", "Builds and refreshes catalog search indexes.", map[string]any{"tier": "async", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/search-indexer", "tags": []string{"search", "batch"}, "business_domain": "catalog"}),
		component("order_fulfillment_worker", kne.KindContainer, "worker", "Order Fulfillment Worker", "Consumes order events and coordinates fulfillment handoff.", map[string]any{"tier": "async", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/order-fulfillment-worker", "tags": []string{"orders", "fulfillment"}, "business_domain": "fulfillment"}),
		component("email_dispatch_worker", kne.KindContainer, "worker", "Email Dispatch Worker", "Sends queued transactional customer email.", map[string]any{"tier": "async", "criticality": "medium", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "platform_team", "repository_external_ref": "rezible-commerce/email-dispatch-worker", "tags": []string{"email", "async"}, "business_domain": "communications"}),
		component("users_postgres", kne.KindContainer, "database", "Users Postgres", "Primary user identity database.", map[string]any{"tier": "data", "criticality": "high", "engine": "postgres", "region": "us-east-1", "owner_team": "identity_team", "tags": []string{"identity", "postgres"}}),
		component("catalog_postgres", kne.KindContainer, "database", "Catalog Postgres", "System of record for product catalog data.", map[string]any{"tier": "data", "criticality": "high", "engine": "postgres", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"catalog", "postgres"}}),
		component("orders_postgres", kne.KindContainer, "database", "Orders Postgres", "System of record for carts, orders, and invoices.", map[string]any{"tier": "data", "criticality": "critical", "engine": "postgres", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"orders", "postgres"}}),
		component("payments_postgres", kne.KindContainer, "database", "Payments Postgres", "Payment transaction and ledger database.", map[string]any{"tier": "data", "criticality": "critical", "engine": "postgres", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"payments", "postgres", "pci"}}),
		component("redis_sessions", kne.KindContainer, "cache", "Redis Sessions", "Session and short-lived identity cache.", map[string]any{"tier": "data", "criticality": "high", "engine": "redis", "region": "us-east-1", "owner_team": "identity_team", "tags": []string{"cache", "sessions"}}),
		component("redis_search_cache", kne.KindContainer, "cache", "Redis Search Cache", "Hot search result and autocomplete cache.", map[string]any{"tier": "data", "criticality": "medium", "engine": "redis", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"cache", "search"}}),
		component("elasticsearch_catalog", kne.KindContainer, "search_cluster", "Elasticsearch Catalog", "Primary product search cluster.", map[string]any{"tier": "data", "criticality": "high", "engine": "elasticsearch", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"search", "index"}}),
		component("s3_invoice_bucket", kne.KindContainer, "object_store", "Invoice S3 Bucket", "Generated invoice PDF storage.", map[string]any{"tier": "data", "criticality": "medium", "engine": "s3", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"invoices", "documents"}}),
		component("order_events_queue", kne.KindContainer, "message_queue", "Order Events Queue", "Durable stream for order lifecycle events.", map[string]any{"tier": "async", "criticality": "high", "engine": "sqs", "region": "us-east-1", "owner_team": "platform_team", "tags": []string{"events", "orders"}}),
		component("stripe", kne.KindSystem, "external_system", "Stripe", "External payment processor.", map[string]any{"tier": "external", "criticality": "critical", "owner_team": "commerce_team", "tags": []string{"payments", "third-party"}}),
		component("sendgrid", kne.KindSystem, "external_system", "SendGrid", "External email delivery provider.", map[string]any{"tier": "external", "criticality": "medium", "owner_team": "platform_team", "tags": []string{"email", "third-party"}}),
		component("customer", kne.KindDomainObject, "business_entity", "Customer", "A person or organization buying from the storefront.", map[string]any{"business_domain": "identity", "owner_team": "identity_team", "tags": []string{"model"}}),
		component("product", kne.KindDomainObject, "business_entity", "Product", "A sellable catalog item.", map[string]any{"business_domain": "catalog", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("cart", kne.KindDomainObject, "business_entity", "Cart", "A customer's active purchase intent.", map[string]any{"business_domain": "checkout", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("order", kne.KindDomainObject, "business_entity", "Order", "A committed customer purchase.", map[string]any{"business_domain": "orders", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("payment", kne.KindDomainObject, "business_entity", "Payment", "Payment authorization and capture record.", map[string]any{"business_domain": "payments", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("invoice", kne.KindDomainObject, "business_entity", "Invoice", "Customer invoice document.", map[string]any{"business_domain": "orders", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("search_index", kne.KindDomainObject, "business_entity", "Search Index", "Materialized product search index.", map[string]any{"business_domain": "catalog", "owner_team": "commerce_team", "tags": []string{"model", "derived"}}),
		component("identity_team", kne.KindActor, "team", "Identity Team", "Owns authentication and customer identity.", map[string]any{"slack_channel": "#team-identity", "oncall_roster": "identity-primary"}),
		component("commerce_team", kne.KindActor, "team", "Commerce Team", "Owns catalog, checkout, orders, and payments.", map[string]any{"slack_channel": "#team-commerce", "oncall_roster": "commerce-primary"}),
		component("platform_team", kne.KindActor, "team", "Platform Team", "Owns shared platform, messaging, and communications infrastructure.", map[string]any{"slack_channel": "#team-platform", "oncall_roster": "platform-primary"}),
	}
}

type topologyRelationshipObservedPayload struct {
	ExternalRef string                                       `json:"external_ref"`
	Kind        knr.Kind                                     `json:"kind"`
	Subkind     string                                       `json:"subkind"`
	DisplayName string                                       `json:"display_name,omitempty"`
	Description string                                       `json:"description,omitempty"`
	Properties  map[string]any                               `json:"properties,omitempty"`
	Source      topologyRelationshipObservedPayloadComponent `json:"source"`
	Target      topologyRelationshipObservedPayloadComponent `json:"target"`
}

type topologyRelationshipObservedPayloadComponent struct {
	ExternalRef string   `json:"external_ref"`
	Kind        kne.Kind `json:"kind"`
	Subkind     string   `json:"subkind"`
	DisplayName string   `json:"display_name"`
}

func (p topologyRelationshipObservedPayload) subjectRef() string {
	return p.ExternalRef
}

func (p topologyRelationshipObservedPayload) getAttributes() projections.SystemRelationshipSubjectAttributes {
	return projections.SystemRelationshipSubjectAttributes{
		ExternalRef:       p.ExternalRef,
		Kind:              p.Kind,
		Subkind:           p.Subkind,
		DisplayName:       p.DisplayName,
		Description:       p.Description,
		SourceExternalRef: p.Source.ExternalRef,
		SourceKind:        p.Source.Kind,
		SourceSubkind:     p.Source.Subkind,
		SourceDisplayName: p.Source.DisplayName,
		TargetExternalRef: p.Target.ExternalRef,
		TargetKind:        p.Target.Kind,
		TargetSubkind:     p.Target.Subkind,
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
					Kind:        c.Kind,
					Subkind:     c.Subkind,
					DisplayName: c.DisplayName,
				}
			}
		}
		panic(fmt.Sprintf("unknown demo topology component: %s", id))
	}

	rel := func(sourceID string, kind knr.Kind, subkind string, targetID string, displayName string) topologyRelationshipObservedPayload {
		source := mustTopologyComponent(sourceID)
		target := mustTopologyComponent(targetID)
		externalRef := fmt.Sprintf("demo:relationship:%s:%s:%s", sourceID, subkind, targetID)
		return topologyRelationshipObservedPayload{
			ExternalRef: externalRef,
			Kind:        kind,
			Subkind:     subkind,
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
		rel("web_app", knr.KindInteractsWith, "calls", "public_api_gateway", "Customer Web App calls Public API Gateway"),
		rel("admin_console", knr.KindInteractsWith, "calls", "public_api_gateway", "Admin Console calls Public API Gateway"),
		rel("public_api_gateway", knr.KindInteractsWith, "calls", "auth_service", "Public API Gateway calls Auth Listener"),
		rel("public_api_gateway", knr.KindInteractsWith, "calls", "catalog_service", "Public API Gateway calls Catalog Listener"),
		rel("public_api_gateway", knr.KindInteractsWith, "calls", "search_api", "Public API Gateway calls Search API"),
		rel("public_api_gateway", knr.KindInteractsWith, "calls", "checkout_service", "Public API Gateway calls Checkout Listener"),
		rel("checkout_service", knr.KindInteractsWith, "calls", "auth_service", "Checkout Listener calls Auth Listener"),
		rel("checkout_service", knr.KindInteractsWith, "calls", "search_api", "Checkout Listener calls Search API"),
		rel("checkout_service", knr.KindInteractsWith, "calls", "inventory_service", "Checkout Listener calls Inventory Listener"),
		rel("checkout_service", knr.KindInteractsWith, "calls", "payments_service", "Checkout Listener calls Payments Listener"),
		rel("checkout_service", knr.KindInteractsWith, "calls", "orders_service", "Checkout Listener calls Orders Listener"),
		rel("orders_service", knr.KindInteractsWith, "calls", "notifications_service", "Orders Listener calls Notifications Listener"),
		rel("auth_service", knr.KindInteractsWith, "reads_from", "redis_sessions", "Auth Listener reads from Redis Sessions"),
		rel("auth_service", knr.KindInteractsWith, "writes_to", "users_postgres", "Auth Listener writes to Users Postgres"),
		rel("catalog_service", knr.KindInteractsWith, "writes_to", "catalog_postgres", "Catalog Listener writes to Catalog Postgres"),
		rel("catalog_service", knr.KindInteractsWith, "publishes_to", "order_events_queue", "Catalog Listener publishes to Order Events Queue"),
		rel("search_api", knr.KindInteractsWith, "reads_from", "elasticsearch_catalog", "Search API reads from Elasticsearch Catalog"),
		rel("search_api", knr.KindInteractsWith, "reads_from", "redis_search_cache", "Search API reads from Redis Search Cache"),
		rel("search_indexer", knr.KindInteractsWith, "reads_from", "catalog_postgres", "Search Indexer reads from Catalog Postgres"),
		rel("search_indexer", knr.KindInteractsWith, "writes_to", "elasticsearch_catalog", "Search Indexer writes to Elasticsearch Catalog"),
		rel("search_indexer", knr.KindInteractsWith, "writes_to", "redis_search_cache", "Search Indexer writes to Redis Search Cache"),
		rel("orders_service", knr.KindInteractsWith, "writes_to", "orders_postgres", "Orders Listener writes to Orders Postgres"),
		rel("orders_service", knr.KindInteractsWith, "publishes_to", "order_events_queue", "Orders Listener publishes to Order Events Queue"),
		rel("payments_service", knr.KindInteractsWith, "writes_to", "payments_postgres", "Payments Listener writes to Payments Postgres"),
		rel("payments_service", knr.KindInteractsWith, "calls", "stripe", "Payments Listener calls Stripe"),
		rel("inventory_service", knr.KindInteractsWith, "reads_from", "catalog_postgres", "Inventory Listener reads from Catalog Postgres"),
		rel("order_fulfillment_worker", knr.KindInteractsWith, "consumes_from", "order_events_queue", "Order Fulfillment Worker consumes from Order Events Queue"),
		rel("email_dispatch_worker", knr.KindInteractsWith, "consumes_from", "order_events_queue", "Email Dispatch Worker consumes from Order Events Queue"),
		rel("notifications_service", knr.KindInteractsWith, "publishes_to", "order_events_queue", "Notifications Listener publishes to Order Events Queue"),
		rel("notifications_service", knr.KindInteractsWith, "calls", "sendgrid", "Notifications Listener calls SendGrid"),
		rel("orders_service", knr.KindInteractsWith, "writes_to", "s3_invoice_bucket", "Orders Listener writes to Invoice S3 Bucket"),
		rel("identity_team", knr.KindOwns, "owns", "auth_service", "Identity Team owns Auth Listener"),
		rel("identity_team", knr.KindOwns, "owns", "users_postgres", "Identity Team owns Users Postgres"),
		rel("commerce_team", knr.KindOwns, "owns", "catalog_service", "Commerce Team owns Catalog Listener"),
		rel("commerce_team", knr.KindOwns, "owns", "search_api", "Commerce Team owns Search API"),
		rel("commerce_team", knr.KindOwns, "owns", "checkout_service", "Commerce Team owns Checkout Listener"),
		rel("commerce_team", knr.KindOwns, "owns", "orders_service", "Commerce Team owns Orders Listener"),
		rel("commerce_team", knr.KindOwns, "owns", "payments_service", "Commerce Team owns Payments Listener"),
		rel("platform_team", knr.KindOwns, "owns", "public_api_gateway", "Platform Team owns Public API Gateway"),
		rel("platform_team", knr.KindOwns, "owns", "notifications_service", "Platform Team owns Notifications Listener"),
		rel("auth_service", knr.KindSupports, "processes_entity", "customer", "Auth Listener processes Customer"),
		rel("catalog_service", knr.KindSupports, "processes_entity", "product", "Catalog Listener processes Product"),
		rel("checkout_service", knr.KindSupports, "processes_entity", "cart", "Checkout Listener processes Cart"),
		rel("orders_service", knr.KindSupports, "processes_entity", "order", "Orders Listener processes Order"),
		rel("payments_service", knr.KindSupports, "processes_entity", "payment", "Payments Listener processes Payment"),
		rel("orders_service", knr.KindSupports, "processes_entity", "invoice", "Orders Listener processes Invoice"),
		rel("search_indexer", knr.KindSupports, "indexes_entity", "product", "Search Indexer indexes Product"),
		rel("elasticsearch_catalog", knr.KindSupports, "stores_entity", "search_index", "Elasticsearch Catalog stores Search Index"),
	}
}
