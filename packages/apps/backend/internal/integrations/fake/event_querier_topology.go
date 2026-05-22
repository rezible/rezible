package fakeprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	rez "github.com/rezible/rezible"
)

const (
	topologyObservationComponent    = "component"
	topologyObservationRelationship = "relationship"
)

type (
	topologyObservedPayload struct {
		ObservationType string                       `json:"observation_type"`
		OccurredAt      time.Time                    `json:"occurred_at"`
		Component       *topologyComponentPayload    `json:"component,omitempty"`
		Relationship    *topologyRelationshipPayload `json:"relationship,omitempty"`
	}

	topologyComponentPayload struct {
		ExternalRef string         `json:"external_ref"`
		Kind        string         `json:"kind"`
		DisplayName string         `json:"display_name"`
		Description string         `json:"description,omitempty"`
		Properties  map[string]any `json:"properties,omitempty"`
	}

	topologyRelationshipPayload struct {
		ExternalRef       string         `json:"external_ref"`
		Kind              string         `json:"kind"`
		DisplayName       string         `json:"display_name,omitempty"`
		Description       string         `json:"description,omitempty"`
		SourceExternalRef string         `json:"source_external_ref"`
		SourceKind        string         `json:"source_kind"`
		SourceDisplayName string         `json:"source_display_name"`
		TargetExternalRef string         `json:"target_external_ref"`
		TargetKind        string         `json:"target_kind"`
		TargetDisplayName string         `json:"target_display_name"`
		Properties        map[string]any `json:"properties,omitempty"`
	}

	fakeTopologyEvent struct {
		Cursor     string
		SubjectRef string
		Payload    topologyObservedPayload
	}
)

func makeFakeTopologyEvents() []fakeTopologyEvent {
	fakeComponents := makeFakeTopologyComponents()
	fakeRels := makeFakeTopologyRelationships(fakeComponents)

	fakeTopologyObservedAt := time.Date(2026, 5, 10, 8, 0, 0, 0, time.UTC)

	events := make([]fakeTopologyEvent, 0, len(fakeComponents)+len(fakeRels))
	for i, payload := range fakeComponents {
		events = append(events, fakeTopologyEvent{
			Cursor:     fmt.Sprintf("component:%03d:%s", i+1, payload.ExternalRef),
			SubjectRef: payload.ExternalRef,
			Payload: topologyObservedPayload{
				ObservationType: topologyObservationComponent,
				OccurredAt:      fakeTopologyObservedAt,
				Component:       &payload,
			},
		})
	}
	for i, payload := range fakeRels {
		payload.Properties["source_external_ref"] = payload.SourceExternalRef
		payload.Properties["target_external_ref"] = payload.TargetExternalRef
		events = append(events, fakeTopologyEvent{
			Cursor:     fmt.Sprintf("relationship:%03d:%s", i+1, payload.ExternalRef),
			SubjectRef: payload.ExternalRef,
			Payload: topologyObservedPayload{
				ObservationType: topologyObservationRelationship,
				OccurredAt:      fakeTopologyObservedAt.Add(time.Minute),
				Relationship:    &payload,
			},
		})
	}
	return events
}

func (q *eventQuerier) pullTopologyEvents(ctx context.Context, cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		for _, fakeEvent := range makeFakeTopologyEvents() {
			if cursor != "" && fakeEvent.Cursor <= cursor {
				continue
			}
			payloadBytes, jsonErr := json.Marshal(fakeEvent.Payload)
			if jsonErr != nil {
				yield(nil, fmt.Errorf("json marshal topology event: %w", jsonErr))
				return
			}
			res := &rez.ProviderEventQueryResult{
				Event: rez.ProviderEvent{
					Provider:         integrationName,
					ProviderSource:   sourceTopology,
					ProviderEventRef: fmt.Sprintf("fake:%s:%s", sourceTopology, fakeEvent.Cursor),
					SubjectRef:       fakeEvent.SubjectRef,
					ReceivedAt:       fakeEvent.Payload.OccurredAt,
					Payload:          payloadBytes,
					ContentType:      "application/json",
				},
				SourceCursorAfter: new(fakeEvent.Cursor),
			}
			if !yield(res, nil) {
				return
			}
		}
	}
}

func componentRef(id string) string {
	return fmt.Sprintf("fake:component:%s", id)
}

func makeFakeTopologyComponents() []topologyComponentPayload {
	component := func(id string, kind string, displayName string, description string, properties map[string]any) topologyComponentPayload {
		props := map[string]any{
			"external_ref": componentRef(id),
		}
		for k, v := range properties {
			props[k] = v
		}
		return topologyComponentPayload{
			ExternalRef: componentRef(id),
			Kind:        kind,
			DisplayName: displayName,
			Description: description,
			Properties:  props,
		}
	}

	return []topologyComponentPayload{
		component("web_app", "user_surface", "Customer Web App", "Primary customer-facing storefront and account experience.", map[string]any{"tier": "edge", "criticality": "high", "lifecycle": "production", "runtime": "sveltekit", "region": "global", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/web-app", "tags": []string{"customer-facing", "frontend"}, "business_domain": "commerce"}),
		component("admin_console", "user_surface", "Admin Console", "Internal operations interface for catalog, search, and order support.", map[string]any{"tier": "internal", "criticality": "medium", "lifecycle": "production", "runtime": "sveltekit", "region": "global", "owner_team": "platform_team", "repository_external_ref": "rezible-commerce/admin-console", "tags": []string{"internal", "operations"}, "business_domain": "operations"}),
		component("public_api_gateway", "gateway", "Public API Gateway", "Ingress gateway for public REST and partner API traffic.", map[string]any{"tier": "edge", "criticality": "high", "lifecycle": "production", "runtime": "envoy", "region": "us-east-1", "owner_team": "platform_team", "tags": []string{"api", "ingress"}, "business_domain": "platform"}),
		component("auth_service", "service", "Auth Service", "Authentication, sessions, and customer identity service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "identity_team", "repository_external_ref": "rezible-commerce/auth-service", "tags": []string{"identity", "sessions"}, "business_domain": "identity"}),
		component("catalog_service", "service", "Catalog Service", "Product catalog read and write API.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/catalog-service", "tags": []string{"products"}, "business_domain": "catalog"}),
		component("search_api", "service", "Search API", "Product search query API used by storefront and checkout enrichment.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/search-api", "tags": []string{"search", "customer-facing"}, "business_domain": "catalog"}),
		component("checkout_service", "service", "Checkout Service", "Cart checkout orchestration and payment initiation.", map[string]any{"tier": "core", "criticality": "critical", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/checkout-service", "tags": []string{"checkout", "revenue"}, "business_domain": "checkout"}),
		component("orders_service", "service", "Orders Service", "Order lifecycle and order history service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/orders-service", "tags": []string{"orders"}, "business_domain": "orders"}),
		component("payments_service", "service", "Payments Service", "Payment capture, refunds, and ledger integration.", map[string]any{"tier": "core", "criticality": "critical", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/payments-service", "tags": []string{"payments", "pci"}, "business_domain": "payments"}),
		component("inventory_service", "service", "Inventory Service", "Stock availability and reservation service.", map[string]any{"tier": "core", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/inventory-service", "tags": []string{"inventory"}, "business_domain": "fulfillment"}),
		component("notifications_service", "service", "Notifications Service", "Customer email and transactional notification API.", map[string]any{"tier": "supporting", "criticality": "medium", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "platform_team", "repository_external_ref": "rezible-commerce/notifications-service", "tags": []string{"email"}, "business_domain": "communications"}),
		component("search_indexer", "worker", "Search Indexer", "Builds and refreshes catalog search indexes.", map[string]any{"tier": "async", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/search-indexer", "tags": []string{"search", "batch"}, "business_domain": "catalog"}),
		component("order_fulfillment_worker", "worker", "Order Fulfillment Worker", "Consumes order events and coordinates fulfillment handoff.", map[string]any{"tier": "async", "criticality": "high", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "commerce_team", "repository_external_ref": "rezible-commerce/order-fulfillment-worker", "tags": []string{"orders", "fulfillment"}, "business_domain": "fulfillment"}),
		component("email_dispatch_worker", "worker", "Email Dispatch Worker", "Sends queued transactional customer email.", map[string]any{"tier": "async", "criticality": "medium", "lifecycle": "production", "runtime": "go", "region": "us-east-1", "owner_team": "platform_team", "repository_external_ref": "rezible-commerce/email-dispatch-worker", "tags": []string{"email", "async"}, "business_domain": "communications"}),
		component("users_postgres", "database", "Users Postgres", "Primary user identity database.", map[string]any{"tier": "data", "criticality": "high", "engine": "postgres", "region": "us-east-1", "owner_team": "identity_team", "tags": []string{"identity", "postgres"}}),
		component("catalog_postgres", "database", "Catalog Postgres", "System of record for product catalog data.", map[string]any{"tier": "data", "criticality": "high", "engine": "postgres", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"catalog", "postgres"}}),
		component("orders_postgres", "database", "Orders Postgres", "System of record for carts, orders, and invoices.", map[string]any{"tier": "data", "criticality": "critical", "engine": "postgres", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"orders", "postgres"}}),
		component("payments_postgres", "database", "Payments Postgres", "Payment transaction and ledger database.", map[string]any{"tier": "data", "criticality": "critical", "engine": "postgres", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"payments", "postgres", "pci"}}),
		component("redis_sessions", "cache", "Redis Sessions", "Session and short-lived identity cache.", map[string]any{"tier": "data", "criticality": "high", "engine": "redis", "region": "us-east-1", "owner_team": "identity_team", "tags": []string{"cache", "sessions"}}),
		component("redis_search_cache", "cache", "Redis Search Cache", "Hot search result and autocomplete cache.", map[string]any{"tier": "data", "criticality": "medium", "engine": "redis", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"cache", "search"}}),
		component("elasticsearch_catalog", "search_cluster", "Elasticsearch Catalog", "Primary product search cluster.", map[string]any{"tier": "data", "criticality": "high", "engine": "elasticsearch", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"search", "index"}}),
		component("s3_invoice_bucket", "object_store", "Invoice S3 Bucket", "Generated invoice PDF storage.", map[string]any{"tier": "data", "criticality": "medium", "engine": "s3", "region": "us-east-1", "owner_team": "commerce_team", "tags": []string{"invoices", "documents"}}),
		component("order_events_queue", "message_queue", "Order Events Queue", "Durable stream for order lifecycle events.", map[string]any{"tier": "async", "criticality": "high", "engine": "sqs", "region": "us-east-1", "owner_team": "platform_team", "tags": []string{"events", "orders"}}),
		component("stripe", "external_system", "Stripe", "External payment processor.", map[string]any{"tier": "external", "criticality": "critical", "owner_team": "commerce_team", "tags": []string{"payments", "third-party"}}),
		component("sendgrid", "external_system", "SendGrid", "External email delivery provider.", map[string]any{"tier": "external", "criticality": "medium", "owner_team": "platform_team", "tags": []string{"email", "third-party"}}),
		component("customer", "business_entity", "Customer", "A person or organization buying from the storefront.", map[string]any{"business_domain": "identity", "owner_team": "identity_team", "tags": []string{"model"}}),
		component("product", "business_entity", "Product", "A sellable catalog item.", map[string]any{"business_domain": "catalog", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("cart", "business_entity", "Cart", "A customer's active purchase intent.", map[string]any{"business_domain": "checkout", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("order", "business_entity", "Order", "A committed customer purchase.", map[string]any{"business_domain": "orders", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("payment", "business_entity", "Payment", "Payment authorization and capture record.", map[string]any{"business_domain": "payments", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("invoice", "business_entity", "Invoice", "Customer invoice document.", map[string]any{"business_domain": "orders", "owner_team": "commerce_team", "tags": []string{"model"}}),
		component("search_index", "business_entity", "Search Index", "Materialized product search index.", map[string]any{"business_domain": "catalog", "owner_team": "commerce_team", "tags": []string{"model", "derived"}}),
		component("identity_team", "team", "Identity Team", "Owns authentication and customer identity.", map[string]any{"slack_channel": "#team-identity", "oncall_roster": "identity-primary"}),
		component("commerce_team", "team", "Commerce Team", "Owns catalog, checkout, orders, and payments.", map[string]any{"slack_channel": "#team-commerce", "oncall_roster": "commerce-primary"}),
		component("platform_team", "team", "Platform Team", "Owns shared platform, messaging, and communications infrastructure.", map[string]any{"slack_channel": "#team-platform", "oncall_roster": "platform-primary"}),
	}
}

func makeFakeTopologyRelationships(cmps []topologyComponentPayload) []topologyRelationshipPayload {
	mustTopologyComponent := func(id string) topologyComponentPayload {
		ref := componentRef(id)
		for _, c := range cmps {
			if c.ExternalRef == ref {
				return c
			}
		}
		panic(fmt.Sprintf("unknown fake topology component: %s", id))
	}

	rel := func(sourceID string, kind string, targetID string, displayName string) topologyRelationshipPayload {
		source := mustTopologyComponent(sourceID)
		target := mustTopologyComponent(targetID)
		externalRef := fmt.Sprintf("fake:relationship:%s:%s:%s", sourceID, kind, targetID)
		return topologyRelationshipPayload{
			ExternalRef:       externalRef,
			Kind:              kind,
			DisplayName:       displayName,
			SourceExternalRef: source.ExternalRef,
			SourceKind:        source.Kind,
			SourceDisplayName: source.DisplayName,
			TargetExternalRef: target.ExternalRef,
			TargetKind:        target.Kind,
			TargetDisplayName: target.DisplayName,
			Properties: map[string]any{
				"external_ref": externalRef,
				"source":       sourceID,
				"target":       targetID,
			},
		}
	}

	return []topologyRelationshipPayload{
		rel("web_app", "calls", "public_api_gateway", "Customer Web App calls Public API Gateway"),
		rel("admin_console", "calls", "public_api_gateway", "Admin Console calls Public API Gateway"),
		rel("public_api_gateway", "calls", "auth_service", "Public API Gateway calls Auth Service"),
		rel("public_api_gateway", "calls", "catalog_service", "Public API Gateway calls Catalog Service"),
		rel("public_api_gateway", "calls", "search_api", "Public API Gateway calls Search API"),
		rel("public_api_gateway", "calls", "checkout_service", "Public API Gateway calls Checkout Service"),
		rel("checkout_service", "calls", "auth_service", "Checkout Service calls Auth Service"),
		rel("checkout_service", "calls", "search_api", "Checkout Service calls Search API"),
		rel("checkout_service", "calls", "inventory_service", "Checkout Service calls Inventory Service"),
		rel("checkout_service", "calls", "payments_service", "Checkout Service calls Payments Service"),
		rel("checkout_service", "calls", "orders_service", "Checkout Service calls Orders Service"),
		rel("orders_service", "calls", "notifications_service", "Orders Service calls Notifications Service"),
		rel("auth_service", "reads_from", "redis_sessions", "Auth Service reads from Redis Sessions"),
		rel("auth_service", "writes_to", "users_postgres", "Auth Service writes to Users Postgres"),
		rel("catalog_service", "writes_to", "catalog_postgres", "Catalog Service writes to Catalog Postgres"),
		rel("catalog_service", "publishes_to", "order_events_queue", "Catalog Service publishes to Order Events Queue"),
		rel("search_api", "reads_from", "elasticsearch_catalog", "Search API reads from Elasticsearch Catalog"),
		rel("search_api", "reads_from", "redis_search_cache", "Search API reads from Redis Search Cache"),
		rel("search_indexer", "reads_from", "catalog_postgres", "Search Indexer reads from Catalog Postgres"),
		rel("search_indexer", "writes_to", "elasticsearch_catalog", "Search Indexer writes to Elasticsearch Catalog"),
		rel("search_indexer", "writes_to", "redis_search_cache", "Search Indexer writes to Redis Search Cache"),
		rel("orders_service", "writes_to", "orders_postgres", "Orders Service writes to Orders Postgres"),
		rel("orders_service", "publishes_to", "order_events_queue", "Orders Service publishes to Order Events Queue"),
		rel("payments_service", "writes_to", "payments_postgres", "Payments Service writes to Payments Postgres"),
		rel("payments_service", "calls", "stripe", "Payments Service calls Stripe"),
		rel("inventory_service", "reads_from", "catalog_postgres", "Inventory Service reads from Catalog Postgres"),
		rel("order_fulfillment_worker", "consumes_from", "order_events_queue", "Order Fulfillment Worker consumes from Order Events Queue"),
		rel("email_dispatch_worker", "consumes_from", "order_events_queue", "Email Dispatch Worker consumes from Order Events Queue"),
		rel("notifications_service", "publishes_to", "order_events_queue", "Notifications Service publishes to Order Events Queue"),
		rel("notifications_service", "calls", "sendgrid", "Notifications Service calls SendGrid"),
		rel("orders_service", "writes_to", "s3_invoice_bucket", "Orders Service writes to Invoice S3 Bucket"),
		rel("identity_team", "owns", "auth_service", "Identity Team owns Auth Service"),
		rel("identity_team", "owns", "users_postgres", "Identity Team owns Users Postgres"),
		rel("commerce_team", "owns", "catalog_service", "Commerce Team owns Catalog Service"),
		rel("commerce_team", "owns", "search_api", "Commerce Team owns Search API"),
		rel("commerce_team", "owns", "checkout_service", "Commerce Team owns Checkout Service"),
		rel("commerce_team", "owns", "orders_service", "Commerce Team owns Orders Service"),
		rel("commerce_team", "owns", "payments_service", "Commerce Team owns Payments Service"),
		rel("platform_team", "owns", "public_api_gateway", "Platform Team owns Public API Gateway"),
		rel("platform_team", "owns", "notifications_service", "Platform Team owns Notifications Service"),
		rel("auth_service", "processes_entity", "customer", "Auth Service processes Customer"),
		rel("catalog_service", "processes_entity", "product", "Catalog Service processes Product"),
		rel("checkout_service", "processes_entity", "cart", "Checkout Service processes Cart"),
		rel("orders_service", "processes_entity", "order", "Orders Service processes Order"),
		rel("payments_service", "processes_entity", "payment", "Payments Service processes Payment"),
		rel("orders_service", "processes_entity", "invoice", "Orders Service processes Invoice"),
		rel("search_indexer", "indexes_entity", "product", "Search Indexer indexes Product"),
		rel("elasticsearch_catalog", "stores_entity", "search_index", "Elasticsearch Catalog stores Search Index"),
	}
}
