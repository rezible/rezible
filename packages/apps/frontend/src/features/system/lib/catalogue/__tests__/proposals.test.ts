import { describe, expect, test } from "bun:test";
import { formatTime } from "$lib/time";
import type {
	KnowledgeGraphEntity,
	KnowledgeGraphEntityAlias,
	KnowledgeProposal,
	KnowledgeProposalValue,
} from "$lib/api";
import {
	confirmQuestion,
	decisionViews,
	entityRowView,
	impliedProposalViews,
	originLabel,
	possibleMatchesByEntity,
	proposalValueViews,
	proposalStatusView,
	proposalView,
	rationaleParts,
	aliasViews,
} from "../proposals";

const now = Date.parse("2026-10-16T12:00:00Z");

const member = { id: "user-1", name: "Ana" };

const valueHeldBy = (value: string, entityId: string | undefined): KnowledgeProposalValue => {
	if (!entityId) {
		return { value };
	}
	return { value, entity: { id: entityId, displayName: value, aliases: [] } };
};

const proposal = (
	status: KnowledgeProposal["attributes"]["status"],
	entityIds: [string | undefined, string | undefined] = ["entity-a", "entity-b"]
): KnowledgeProposal => ({
	id: `proposal-${status}`,
	attributes: {
		category: "container",
		kind: "service",
		attribute: "service.name",
		values: [valueHeldBy("checkout", entityIds[0]), valueHeldBy("checkout-api", entityIds[1])],
		status,
		origin: "name_similarity",
		rationale: "They differ only by a generic term.",
		evidence: {
			id: "evidence-1",
			integrationDisplayName: "Webhook",
			provider: "webhook",
			resourceRef: "service:checkout",
			assertion: "deployment_service_observed",
			effectiveAt: "2026-10-16T11:00:00Z",
			displayName: "checkout",
		},
		decisions: [],
		createdAt: "2026-10-16T11:00:00Z",
	},
});

describe("proposal status", () => {
	test("a pending proposal can be confirmed or rejected", () => {
		expect(proposalStatusView("pending")).toEqual({
			label: "Proposed match",
			tone: "pending",
			actions: ["confirm", "reject"],
		});
	});

	test("a rejected proposal can still be confirmed", () => {
		expect(proposalStatusView("rejected")).toMatchObject({ tone: "rejected", actions: ["confirm"] });
	});

	test("a confirmed proposal is final", () => {
		expect(proposalStatusView("confirmed")).toMatchObject({ tone: "confirmed", actions: [] });
	});
});

describe("decisions", () => {
	test("say who decided, and when a confirmation came with another match", () => {
		const views = decisionViews(
			[
				{ status: "rejected", user: member, createdAt: "2026-10-16T10:00:00Z" },
				{
					status: "confirmed",
					user: member,
					createdAt: "2026-10-16T11:00:00Z",
					impliedByProposalId: "other",
				},
			],
			now
		);

		expect(views.map((view) => view.text)).toEqual([
			"Rejected by Ana",
			"Confirmed by Ana with another match",
		]);
		expect(views[0].time).toBe("2 hours ago");
	});
});

describe("aliases", () => {
	test("say when Rezible last recorded them, and whether a confirmed match brought them", () => {
		const aliases: KnowledgeGraphEntityAlias[] = [
			{
				id: "alias-1",
				attributes: {
					kind: "entity",
					resourceRef: { provider: "webhook", providerNamespace: "", resourceRef: "checkout" },
				},
				integrationDisplayName: "",
				provider: "webhook",
				name: "checkout",
				lastRecordedAt: "2026-10-16T11:55:00Z",
				mergedByProposal: {
					proposalId: "proposal-1",
					confirmedBy: member,
					confirmedAt: "2026-10-16T11:00:00Z",
				},
			},
			{
				id: "alias-2",
				attributes: {
					kind: "entity",
					resourceRef: { provider: "webhook", providerNamespace: "", resourceRef: "checkout-api" },
				},
				integrationDisplayName: "Monitoring",
				provider: "monitoring",
				name: "checkout-api",
			},
		];

		const [merged, unrecorded] = aliasViews(aliases, now);

		expect(merged.integration).toBe("webhook");
		expect(merged.lastRecorded).toBe(`last recorded ${formatTime("2026-10-16T11:55:00Z").absolute}`);
		expect(merged.mergedNote).toContain("Ana");
		expect(unrecorded.lastRecorded).toBe("nothing recorded");
		expect(unrecorded.mergedNote).toBeUndefined();
	});
});

describe("values", () => {
	test("a value no entity holds any more is marked", () => {
		const [held, unheld] = proposalValueViews(
			proposal("pending", ["entity-a", undefined]).attributes.values,
			now
		);

		expect(held.unheld).toBe(false);
		expect(held.displayName).toBeUndefined();
		expect(unheld.unheld).toBe(true);
	});
});

describe("confirming", () => {
	test("asks about both values as one entity of the proposal's kind", () => {
		expect(confirmQuestion(proposal("pending")).title).toBe(
			"Treat checkout and checkout-api as one service?"
		);
	});

	test("lists the proposals it settles, with who rejected a rejected one", () => {
		const views = impliedProposalViews(
			[
				{ id: "p1", values: ["checkout", "checkout-svc"], status: "pending" },
				{
					id: "p2",
					values: ["checkout-api", "checkout-svc"],
					status: "rejected",
					rejectedBy: member,
					rejectedAt: "2026-10-15T12:00:00Z",
				},
			],
			now
		);

		expect(views[0]).toEqual({ key: "p1", values: "checkout and checkout-svc" });
		expect(views[1].rejection).toBe("rejected by Ana, yesterday");
	});
});

describe("possible matches", () => {
	test("map each entity a proposal names to it, the last listed winning", () => {
		const older = { ...proposal("pending"), id: "older" };
		const newer = { ...proposal("pending", ["entity-a", "entity-c"]), id: "newer" };

		const matches = possibleMatchesByEntity([newer, older]);

		expect(matches.get("entity-a")).toBe("older");
		expect(matches.get("entity-b")).toBe("older");
		expect(matches.get("entity-c")).toBe("newer");
	});
});

describe("entity rows", () => {
	const entity = (displayName: string, values: string[]): KnowledgeGraphEntity => ({
		id: "entity-a",
		attributes: {
			category: "container",
			kind: "service",
			displayName,
			linkingAttributes: values.map((value) => ({ attribute: "service.name", value })),
			aliases: [],
			createdAt: "2026-10-16T11:00:00Z",
			updatedAt: "2026-10-16T11:00:00Z",
		},
	});

	test("fall back to the first value, then to Unnamed, without a display name", () => {
		const matches = new Map([["entity-a", "proposal-1"]]);

		expect(entityRowView(entity("Checkout", ["checkout"]), matches, now).displayName).toBe("Checkout");
		expect(entityRowView(entity("", ["checkout", "checkout-api"]), matches, now).displayName).toBe(
			"checkout"
		);
		expect(entityRowView(entity("", []), new Map(), now).displayName).toBe("Unnamed");
		expect(entityRowView(entity("", []), matches, now).possibleMatchProposalId).toBe("proposal-1");
	});
});

describe("origin", () => {
	test("says why Rezible proposed the match", () => {
		expect(originLabel("name_similarity")).toBe("Similar names");
		expect(originLabel("identity_conflict")).toBe("A source names it differently");
	});
});

describe("proposal view", () => {
	test("shows the entity once when both values already belong to it", () => {
		expect(proposalView(proposal("confirmed", ["entity-a", "entity-a"]), now).sharedEntity).toEqual({
			kind: "service",
			aliases: [],
		});
		expect(proposalView(proposal("pending"), now).sharedEntity).toBeUndefined();
	});

	test("shows the values in the rationale as code", () => {
		expect(rationaleParts("`checkout` and `checkout-api` differ")).toEqual([
			{ key: "1", text: "checkout", code: true },
			{ key: "2", text: " and ", code: false },
			{ key: "3", text: "checkout-api", code: true },
			{ key: "4", text: " differ", code: false },
		]);
	});
});
