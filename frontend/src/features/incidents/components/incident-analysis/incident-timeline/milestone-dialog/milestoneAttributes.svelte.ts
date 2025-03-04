import type { Incident, IncidentMilestoneAttributes } from "$lib/api";
import { createMentionEditor } from "$features/incidents/lib/editor.svelte";
import type { Content } from "@tiptap/core";
import {now, getLocalTimeZone, type ZonedDateTime, parseAbsoluteToLocal} from '@internationalized/date';

const makeTimeAnchor = (from?: string): ZonedDateTime => {
	if (from) return parseAbsoluteToLocal(from);
	return now(getLocalTimeZone());
};

type DescriptionEditor = ReturnType<typeof createMentionEditor> | null;
type MilestoneKind = IncidentMilestoneAttributes["kind"];
const createMilestoneAttributesState = () => {
	let title = $state<string>("");
	let kind = $state<MilestoneKind>("impact");
	let descriptionContent = $state<Content>();
	let descriptionEditor = $state<DescriptionEditor>(null);
	let timestamp = $state<ZonedDateTime>(makeTimeAnchor());

	const valid = $derived(true);

	const init = (inc?: Incident, e?: IncidentMilestoneAttributes) => {
		title = $state.snapshot(e?.title) ?? "";
		// descriptionContent = (!!e?.description) ? JSON.parse(e.description) as Content : undefined;
		timestamp = makeTimeAnchor(e?.timestamp ?? inc?.attributes.openedAt);
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

	return {
		init,
		get timestamp() { return timestamp },
		set timestamp(t: ZonedDateTime) { timestamp = t; onUpdate(); },
		get title() { return title },
		set title(t: string) { title = t; onUpdate(); },
		get kind() { return kind },
		set kind(k: MilestoneKind) { kind = k; onUpdate(); },
		mountDescriptionEditor,
		get descriptionEditor() { return descriptionEditor },

		get valid() { return valid },
		
		snapshot(): IncidentMilestoneAttributes {
			return $state.snapshot({
				incidentId: "",
				title,
				kind,
				description: getDescriptionContent(),
				timestamp: timestamp.toAbsoluteString(),
				type: "",
			})
		},
	}
}

export const milestoneAttributes = createMilestoneAttributesState();