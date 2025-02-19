import type { IncidentEventDecisionContext, DateTimeAnchor, IncidentEventAttributes, IncidentEventContributingFactor, IncidentEventEvidence, IncidentEventSystemComponent, Incident } from "$lib/api";
import { createMentionEditor } from "$features/incidents/lib/editor.svelte";
import type { Content, JSONContent } from "@tiptap/core";
import { convertDateTimeAnchor } from "$lib/utils.svelte";

const makeTimeAnchor = (from?: Date) => {
	if (!from || true) return {
		date: new Date(),
		time: "09:00:00",
		timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
	}

};

const makeDefaultDecisionContext = () => ({
	optionsConsidered: [],
	constraints: [],
	decisionRationale: "",
});

type DescriptionEditor = ReturnType<typeof createMentionEditor> | null;
type EventKind = IncidentEventAttributes["kind"];
const createEventAttributesState = () => {
	let kind = $state<EventKind>("observation");
	let title = $state<string>("");
	let descriptionContent = $state<Content>();
	let descriptionEditor = $state<DescriptionEditor>(null);
	let timestamp = $state<DateTimeAnchor>(makeTimeAnchor());
	let isKey = $state(false);
	let decisionContext = $state<IncidentEventDecisionContext>(makeDefaultDecisionContext());
	let contributingFactors = $state<IncidentEventContributingFactor[]>([]);
	let evidence = $state<IncidentEventEvidence[]>([]);
	let systemContext = $state<IncidentEventSystemComponent[]>([]);

	const valid = $derived(true);

	const init = (inc?: Incident, e?: IncidentEventAttributes) => {
		kind = $state.snapshot(e?.kind) ?? "observation";
		title = $state.snapshot(e?.title) ?? "";
		descriptionContent = (!!e?.description) ? JSON.parse(e.description) as Content : undefined;
		isKey = $state.snapshot(e?.isKey) ?? false;
		const initTime = $state.snapshot(e?.timestamp) ?? inc?.attributes.openedAt; // TODO: use incident start time
		timestamp = makeTimeAnchor(initTime);
		decisionContext = $state.snapshot(e?.decisionContext) ?? makeDefaultDecisionContext();
		contributingFactors = $state.snapshot(e?.contributingFactors) ?? [];
		evidence = $state.snapshot(e?.evidence) ?? [];
		systemContext = $state.snapshot(e?.systemContext) ?? [];
	}

	const onUpdate = () => {
		// TODO: check if attributes valid;
	}

	const mountDescriptionEditor = () => {
		descriptionEditor = createMentionEditor(descriptionContent ?? "", "cursor-text focus:outline-none min-h-20");
		// TODO: watch for update?
		return () => {
			descriptionEditor?.destroy();
			descriptionEditor = null;
		}
	}

	const getDescriptionContent = () => {
		if (!descriptionEditor) return;
		return JSON.stringify(descriptionEditor.getJSON());
	}

	// this is gross but oh well
	return {
		init,
		get kind() { return kind },
		set kind(k: EventKind) { kind = k; onUpdate(); },
		get timestamp() { return timestamp },
		set timestamp(t: DateTimeAnchor) { timestamp = t; onUpdate(); },
		get isKey() { return isKey },
		set isKey(v: boolean) { isKey = v; onUpdate() },
		get title() { return title },
		set title(t: string) { title = t; onUpdate(); },
		mountDescriptionEditor,
		get descriptionEditor() { return descriptionEditor },
		get decisionContext() { return decisionContext },
		set decisionContext(dc: IncidentEventDecisionContext) { decisionContext = dc; onUpdate(); },
		get contributingFactors() { return contributingFactors },
		set contributingFactors(cf: IncidentEventContributingFactor[]) { contributingFactors = cf; onUpdate(); },
		get evidence() { return evidence },
		set evidence(e: IncidentEventEvidence[]) { evidence = e; onUpdate(); },
		get systemContext() { return systemContext },
		set systemContext(sc: IncidentEventSystemComponent[]) { systemContext = sc; onUpdate(); },

		get valid() { return valid },
		
		snapshot(): IncidentEventAttributes {
			return {
				incidentId: "",
				sequence: -1,
				kind: $state.snapshot(kind),
				title: $state.snapshot(title),
				description: getDescriptionContent(),
				timestamp: convertDateTimeAnchor($state.snapshot(timestamp)),
				isKey: $state.snapshot(isKey),
				decisionContext: $state.snapshot(decisionContext),
				contributingFactors: $state.snapshot(contributingFactors),
				evidence: $state.snapshot(evidence),
				systemContext: $state.snapshot(systemContext),
			}
		},
	}
}

export const eventAttributes = createEventAttributesState();