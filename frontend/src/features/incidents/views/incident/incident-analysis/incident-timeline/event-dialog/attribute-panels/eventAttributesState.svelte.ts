import type { IncidentEventDecisionContext, IncidentEventAttributes, IncidentEventContributingFactor, IncidentEventEvidence, IncidentEventSystemComponent, Incident } from "$lib/api";
import { createMentionEditor } from "$components/tiptap-editor/editors";
import type { Content } from "@tiptap/core";
import {now, getLocalTimeZone, type ZonedDateTime, parseAbsoluteToLocal} from '@internationalized/date';

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

export class TimelineEventDialogAttributesState {
	kind = $state<EventKind>("observation");
	title = $state<string>("");
	descriptionContent = $state<Content>();
	descriptionEditor = $state<DescriptionEditor>(null);
	timestamp = $state<ZonedDateTime>(makeTimeAnchor());
	isKey = $state(false);
	decisionContext = $state<IncidentEventDecisionContext>(makeDefaultDecisionContext());
	contributingFactors = $state<IncidentEventContributingFactor[]>([]);
	evidence = $state<IncidentEventEvidence[]>([]);
	systemContext = $state<IncidentEventSystemComponent[]>([]);

	init(inc?: Incident, e?: IncidentEventAttributes) {
		this.kind = $state.snapshot(e?.kind) ?? "observation";
		this.title = $state.snapshot(e?.title) ?? "";
		this.descriptionContent = (!!e?.description) ? JSON.parse(e.description) as Content : undefined;
		this.isKey = $state.snapshot(e?.isKey) ?? false;
		this.timestamp = makeTimeAnchor(e?.timestamp ?? inc?.attributes.openedAt); // TODO: use incident start time
		this.decisionContext = $state.snapshot(e?.decisionContext) ?? makeDefaultDecisionContext();
		this.contributingFactors = $state.snapshot(e?.contributingFactors) ?? [];
		this.evidence = $state.snapshot(e?.evidence) ?? [];
		this.systemContext = $state.snapshot(e?.systemContext) ?? [];
	}

	onUpdate() {
		// TODO: check if attributes valid;
	}

	mountDescriptionEditor() {
		this.descriptionEditor = createMentionEditor(this.descriptionContent ?? "", "cursor-text focus:outline-none min-h-20");
		return () => {
			this.descriptionEditor?.destroy();
			this.descriptionEditor = null;
		}
	}

	getDescriptionContent() {
		if (!this.descriptionEditor) return;
		return JSON.stringify(this.descriptionEditor.getJSON());
	}

	snapshot() {
		return $state.snapshot({
			kind: this.kind,
			title: this.title,
			description: this.getDescriptionContent(),
			timestamp: this.timestamp.toAbsoluteString(),
			isKey: this.isKey,
			decisionContext: this.decisionContext,
			contributingFactors: this.contributingFactors,
			evidence: this.evidence,
			systemContext: this.systemContext,
		})
	}
}

export const eventAttributes = new TimelineEventDialogAttributesState();