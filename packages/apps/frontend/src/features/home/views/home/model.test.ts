import { expect, mock, test } from "bun:test";
import type { ActivityRecord, InboxItem, Incident, Situation } from "$lib/api";

// SvelteKit's $app modules are unavailable under bun; resolve route IDs literally.
mock.module("$app/paths", () => ({
	resolve: (route: string, params: Record<string, string | undefined> = {}) =>
		route
			.replace(/\/\[\[(\w+)(=\w+)?\]\]/g, (_, name) => (params[name] ? `/${params[name]}` : ""))
			.replace(/\[(\w+)(=\w+)?\]/g, (_, name) => params[name] ?? ""),
}));

const { activityHref, activityRow, inboxItemHref, inboxRow, incidentRow, situationRow } =
	await import("./model");

const now = Date.parse("2026-05-14T12:00:00Z");

const inboxItem = (attributes: Partial<InboxItem["attributes"]>): InboxItem => ({
	id: "item",
	attributes: {
		context: "",
		kind: "task",
		occurredAt: "2026-05-14T10:00:00Z",
		reason: "Assigned to you",
		state: "open",
		targetKind: "task",
		title: "Follow up",
		...attributes,
	},
});

const activity = (attributes: Partial<ActivityRecord["attributes"]>): ActivityRecord => ({
	id: "activity",
	attributes: {
		explanation: "",
		occurredAt: "2026-05-14T11:00:00Z",
		recordId: "record",
		recordKind: "incident-update",
		scope: "team",
		title: "Update",
		...attributes,
	},
});

test("inboxItemHref maps each target kind", () => {
	expect(inboxItemHref(inboxItem({ targetKind: "task", targetId: "t1" }))).toBe("/tasks/t1");
	expect(inboxItemHref(inboxItem({ targetKind: "task" }))).toBeUndefined();
	expect(inboxItemHref(inboxItem({ targetKind: "normalized-event", targetId: "e1" }))).toBe("/events/e1");
	expect(inboxItemHref(inboxItem({ targetKind: "discussion-thread", incidentId: "i1" }))).toBe(
		"/incidents/i1/report"
	);
	expect(inboxItemHref(inboxItem({ targetKind: "discussion-thread" }))).toBeUndefined();
	expect(inboxItemHref(inboxItem({ targetKind: "maintenance-request", targetId: "m1" }))).toBeUndefined();
});

test("inbox rows only offer an action with a destination", () => {
	expect(inboxRow(inboxItem({ targetKind: "task", targetId: "t1" }), now).actionLabel).toBe("Open task");
	expect(inboxRow(inboxItem({ targetKind: "maintenance-request" }), now).actionLabel).toBeUndefined();
});

test("activityHref maps each record kind", () => {
	expect(activityHref(activity({ recordKind: "incident-update", incidentId: "i1" }))).toBe("/incidents/i1");
	expect(activityHref(activity({ recordKind: "incident-update" }))).toBeUndefined();
	expect(activityHref(activity({ recordKind: "situation-investigation", situationId: "s1" }))).toBe(
		"/situations/s1/investigation"
	);
	expect(activityHref(activity({ recordKind: "inbox-item" }))).toBeUndefined();
});

test("due status distinguishes overdue from upcoming", () => {
	expect(inboxRow(inboxItem({ dueAt: "2026-05-14T11:00:00Z" }), now).due?.label).toBe("Overdue");
	const upcoming = inboxRow(inboxItem({ dueAt: "2026-05-14T15:00:00Z" }), now).due;
	expect(upcoming?.label).toContain("Due");
	expect(upcoming?.tone).toBe("neutral");
	expect(inboxRow(inboxItem({}), now).due).toBeUndefined();
});

test("incident rows dedupe and limit service names", () => {
	const impact = (name: string | undefined) => ({
		id: name ?? "none",
		note: "",
		source: "",
		knowledgeEntity: {
			id: "entity",
			attributes: name ? { latestState: { displayName: name } } : undefined,
		},
	});
	const incident = {
		id: "incident",
		attributes: {
			slug: "inc",
			title: "Incident",
			summary: " Summary ",
			responseState: "started",
			openedAt: "2026-05-14T10:00:00Z",
			resolvedAt: null,
			updatedAt: "2026-05-14T11:00:00Z",
			impacts: ["a", "b", "a", undefined, "c", "d"].map(impact),
		},
	} as unknown as Incident;

	const row = incidentRow(incident);
	expect(row.services).toEqual(["a", "b", "c"]);
	expect(row.moreServices).toBe(1);
	expect(row.summary).toBe("Summary");
	expect(row.response.label).toBe("Active");
});

test("situation source count dedupes sources within groups", () => {
	const situation = {
		id: "situation",
		attributes: {
			title: "Situation",
			summary: "",
			openedAt: "2026-05-14T10:00:00Z",
			linkedIncidentIds: ["i1"],
			observationGroups: [
				{
					id: "g1",
					attributes: { events: [{ id: "e1" }, { id: "e1" }], alertEpisodes: [{ id: "a1" }] },
				},
				{ id: "g2", attributes: { events: [{ id: "e2" }], alertEpisodes: [] } },
			],
		},
	} as unknown as Situation;

	const row = situationRow(situation, { kind: "none" });
	expect(row.sourceCount).toBe(3);
	expect(row.linkedIncidentCount).toBe(1);
	expect(row.href).toBe("/situations/situation");
});

test("activity rows group by local day", () => {
	expect(activityRow(activity({ occurredAt: new Date(now - 60_000).toISOString() }), now).day).toBe(
		"Today"
	);
	expect(activityRow(activity({ occurredAt: "2026-05-01T00:00:00Z" }), now).day).toBe("Earlier");
});
