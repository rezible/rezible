import { expect, test } from "bun:test";
import type {
	InvestigationProviderRead,
	InvestigationReportAttributes,
	SituationBrief,
	SituationBriefDeployment,
	SituationBriefService,
} from "$lib/api";
import {
	briefSections,
	changesState,
	checkedState,
	providerReadResultView,
	unseenState,
	type BriefLoad,
} from "./brief";

const service = (
	id: string,
	name: string,
	overrides: Partial<SituationBriefService> = {}
): SituationBriefService => ({
	entityId: id,
	displayName: name,
	unconfirmedMatches: [],
	...overrides,
});

const deployment = (
	id: string,
	overrides: Partial<SituationBriefDeployment> = {}
): SituationBriefDeployment => ({
	id,
	deployedAt: "2026-10-15T09:30:00Z",
	afterOpened: false,
	environment: "production",
	status: "succeeded",
	sha: "4f1c0d9e8b7a6f5e4d3c",
	version: "",
	url: "",
	repositoryName: "acme/checkout",
	mergedChangesUnlinked: "",
	mergedChanges: [],
	mergedChangesCut: false,
	...overrides,
});

const brief = (overrides: Partial<SituationBrief> = {}): SituationBrief => ({
	services: [service("checkout", "checkout-api")],
	changes: {
		available: true,
		unavailableReason: "",
		start: "2026-10-14T10:00:00Z",
		openedAt: "2026-10-15T10:00:00Z",
		end: "2026-10-15T12:00:00Z",
		services: [{ serviceId: "checkout", deployments: [], limits: [] }],
	},
	unseen: [{ kind: "not_read", text: "Rezible does not read metrics or traces yet." }],
	...overrides,
});

const loaded = (value: SituationBrief): BriefLoad => ({ kind: "loaded", brief: value });

const report = (overrides: Partial<InvestigationReportAttributes> = {}) =>
	({
		text: "",
		summary: "",
		nextChecks: [],
		references: [],
		...overrides,
	}) as InvestigationReportAttributes;

test("briefSections lists the interpretation, signals, changes, what was checked and what cannot be seen, in order", () => {
	const sections = briefSections({ kind: "loading" }, undefined);

	expect(sections.map((section) => section.kind)).toEqual([
		"interpretation",
		"signals",
		"changes",
		"checked",
		"unseen",
	]);
});

test("the interpretation shows the latest report's next checks, at most three", () => {
	const nextChecks = [
		{ check: "Open the checkout dashboard", reason: "Rezible does not read metrics." },
		{ check: "Ask the payments on-call", reason: "Its timeouts started first." },
		{ check: "Check the pool size", reason: "It changed in the deployment." },
		{ check: "A fourth", reason: "Never shown." },
	];

	const [interpretation] = briefSections(loaded(brief()), report({ nextChecks }));
	const [withoutReport] = briefSections(loaded(brief()), undefined);
	const [withoutChecks] = briefSections(loaded(brief()), report());

	expect(
		interpretation.kind === "interpretation" && interpretation.nextChecks.map((next) => next.check)
	).toEqual(["Open the checkout dashboard", "Ask the payments on-call", "Check the pool size"]);
	expect(withoutReport.kind === "interpretation" && withoutReport.nextChecks).toEqual([]);
	expect(withoutChecks.kind === "interpretation" && withoutChecks.nextChecks).toEqual([]);
});

test("changesState: loading and a failed load", () => {
	expect(changesState({ kind: "loading" })).toEqual({ kind: "loading" });
	expect(changesState({ kind: "error" })).toEqual({ kind: "error" });
});

test("changesState: a situation without services says so", () => {
	const state = changesState(loaded(brief({ services: [] })));

	expect(state).toEqual({ kind: "no_services" });
});

test("changesState: no integration reports deployments", () => {
	const value = brief();
	value.changes = {
		...value.changes,
		available: false,
		unavailableReason: "no deployments integration installed",
		services: [],
	};

	const state = changesState(loaded(value));

	expect(state.kind).toBe("not_reported");
});

test("changesState: nothing found says over which window", () => {
	const state = changesState(loaded(brief()));

	expect(state.kind).toBe("nothing_found");
	if (state.kind === "nothing_found") {
		expect(state.window.start).toBe("2026-10-14T10:00:00Z");
		expect(state.window.end).toBe("2026-10-15T12:00:00Z");
		expect(state.services.map((listed) => listed.name)).toEqual(["checkout-api"]);
	}
});

test("changesState: deployments link to their source, else their evidence, and say why changes are unlinked", () => {
	const value = brief({
		services: [
			service("checkout", "checkout-api", {
				unconfirmedMatches: [
					{
						proposalId: "proposal",
						name: "checkout",
						entityId: "other",
						displayName: "checkout",
						rationale: "`checkout` and `checkout-api` differ only by the generic term `api`",
					},
				],
			}),
			service("payments", "payments-api"),
		],
	});
	value.changes.services = [
		{
			serviceId: "checkout",
			deployments: [
				deployment("after", {
					afterOpened: true,
					url: "https://ci.example/runs/2",
					version: "checkout-api@2042",
				}),
				deployment("before", {
					latestEvidenceId: "evidence",
					mergedChanges: [
						{
							id: "change",
							number: 42,
							title: "Lower the payments timeout",
							author: "alice",
							mergedAt: "2026-10-15T09:00:00Z",
							url: "https://forge.example/acme/checkout/pull/42",
							link: "merge_commit",
						},
					],
				}),
				deployment("unlinked", { mergedChangesUnlinked: "no_repository" }),
			],
			limits: [],
		},
		{ serviceId: "payments", deployments: [], limits: [] },
	];

	const state = changesState(loaded(value));

	expect(state.kind).toBe("deployments");
	if (state.kind !== "deployments") {
		return;
	}
	const [checkout, payments] = state.services;
	expect(checkout.unconfirmedMatches.map((match) => match.label)).toEqual([
		"May also be checkout (unconfirmed)",
	]);
	const [after, before, unlinked] = checkout.deployments;
	expect(after.afterOpened).toBe(true);
	expect(after.title).toBe("checkout-api@2042");
	expect(after.link).toEqual({ kind: "url", href: "https://ci.example/runs/2" });
	expect(before.link).toEqual({ kind: "evidence", evidenceId: "evidence" });
	expect(before.mergedChanges[0].detail).toBe("#42 by alice");
	expect(before.unlinkedReason).toBeUndefined();
	expect(unlinked.link).toBeUndefined();
	expect(unlinked.unlinkedReason).toBeDefined();
	expect(payments.deployments).toEqual([]);
});

test("checkedState lists the latest report's provider reads in citation order", () => {
	const summary = {
		toolName: "search_logs",
		providerDisplayName: "Logs",
		serviceName: "checkout-api",
		start: "2026-10-15T09:00:00Z",
		end: "2026-10-15T10:00:00Z",
		query: '{service="checkout-api"}',
		limits: [],
		exploreUrl: "",
		status: "succeeded" as const,
	};
	const withReads = report({
		references: [
			{ id: "evidence", kind: "knowledge_evidence" },
			{ id: "read", kind: "provider_read", providerRead: summary },
		],
	});

	expect(checkedState(undefined)).toEqual({ kind: "no_report" });
	expect(checkedState(report({ references: [{ id: "evidence", kind: "knowledge_evidence" }] }))).toEqual({
		kind: "none_cited",
	});
	const state = checkedState(withReads);
	expect(state.kind === "reads" && state.reads.map((read) => [read.id, read.number])).toEqual([
		["read", 2],
	]);
});

test("unseenState lists every entry, and is unavailable when the brief is", () => {
	const value = brief({
		unseen: [
			{ kind: "source_unavailable", text: "Logs cannot be read: no logs integration installed." },
			{ kind: "not_read", text: "Rezible does not read metrics or traces yet." },
		],
	});

	const state = unseenState(loaded(value));

	expect(state.kind === "entries" && state.entries.map((entry) => entry.kind)).toEqual([
		"source_unavailable",
		"not_read",
	]);
	expect(unseenState({ kind: "error" })).toEqual({ kind: "error" });
});

test("providerReadResultView: lines newest first, counts as bars, failed reads without their error", () => {
	const read = (attributes: Partial<InvestigationProviderRead["attributes"]>, status = "succeeded") =>
		({
			id: "read",
			attributes: {
				summary: { start: "2026-10-15T09:00:00Z", end: "2026-10-15T10:00:00Z", status },
				...attributes,
			},
		}) as InvestigationProviderRead;

	const search = providerReadResultView(
		read({
			logSearch: {
				labels: { service: "checkout-api" },
				lines: [
					{ time: "2026-10-15T09:50:00Z", line: "retrying", labels: {}, count: 1 },
					{
						time: "2026-10-15T09:59:00Z",
						line: "upstream timeout",
						labels: { pod: "a" },
						count: 3,
					},
				],
			},
		})
	);
	expect(search.kind === "log_search" && search.lines.map((line) => [line.line, line.repeats])).toEqual([
		["upstream timeout", "×3"],
		["retrying", undefined],
	]);

	const count = providerReadResultView(
		read({
			logCount: {
				points: [
					{ time: "2026-10-15T09:55:00Z", value: 1 },
					{ time: "2026-10-15T10:00:00Z", value: 100 },
					{ time: "2026-10-15T10:05:00Z", value: 0 },
				],
				stepSeconds: 300,
			},
		})
	);
	expect(count.kind === "log_count" && count.points.map((point) => point.heightPercent)).toEqual([
		4, 100, 0,
	]);
	expect(count.kind === "log_count" && count.step).toBe("5m");

	expect(providerReadResultView(read({}, "failed"))).toEqual({ kind: "failed" });
	expect(providerReadResultView(read({}))).toEqual({ kind: "unavailable" });
});
