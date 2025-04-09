import type { IncidentEventDecisionContext, IncidentEventAttributes, IncidentEventContributingFactor, IncidentEventEvidence, IncidentEventSystemComponent, Incident } from "$lib/api";
import { createMentionEditor } from "$features/incidents/lib/editor.svelte";
import type { Content, JSONContent } from "@tiptap/core";
import {now, getLocalTimeZone, type ZonedDateTime, fromDate, parseAbsolute, parseAbsoluteToLocal} from '@internationalized/date';

const makeTimeAnchor = (from?: string): ZonedDateTime => {
	if (from) return parseAbsoluteToLocal(from);
	return now(getLocalTimeZone());
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
	let timestamp = $state<ZonedDateTime>(makeTimeAnchor());
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
		timestamp = makeTimeAnchor(e?.timestamp ?? inc?.attributes.openedAt); // TODO: use incident start time
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
		set timestamp(t: ZonedDateTime) { timestamp = t; onUpdate(); },
		get isKey() { return isKey },
		set isKey(v: boolean) { isKey = v; onUpdate() },
		get title() { return title },
		set title(t: string) { title = t; onUpdate(); },
		mountDescriptionEditor,
		get descriptionEditor() { return descriptionEditor },
		set descriptionEditor(editor: DescriptionEditor) { descriptionEditor = editor },
		get decisionContext() { return decisionContext },
		set decisionContext(dc: IncidentEventDecisionContext) { decisionContext = dc; onUpdate(); },
		get contributingFactors() { return contributingFactors },
		set contributingFactors(cf: IncidentEventContributingFactor[]) { contributingFactors = cf; onUpdate(); },
		get evidence() { return evidence },
		set evidence(e: IncidentEventEvidence[]) { evidence = e; onUpdate(); },
		get systemContext() { return systemContext },
		set systemContext(sc: IncidentEventSystemComponent[]) { systemContext = sc; onUpdate(); },

		get valid() { return valid },
		
		snapshot() {
			return $state.snapshot({
				kind,
				title,
				description: getDescriptionContent(),
				timestamp: timestamp.toAbsoluteString(),
				isKey,
				decisionContext,
				contributingFactors,
				evidence,
				systemContext,
			})
		},
	}
}

export const eventAttributes = createEventAttributesState();