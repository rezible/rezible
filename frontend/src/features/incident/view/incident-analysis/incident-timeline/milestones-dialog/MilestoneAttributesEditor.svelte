<script lang="ts">
	import { Field, Tooltip, ToggleGroup, ToggleOption } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiShape } from "@mdi/js";
	import TiptapEditor from "$components/tiptap-editor/TiptapEditor.svelte";
	import { createMutation } from "@tanstack/svelte-query";
	import { onMount } from "svelte";
	import {
		createIncidentMilestoneMutation,
		updateIncidentMilestoneMutation,
		type CreateIncidentMilestoneAttributes,
		type IncidentMilestone,
		type IncidentMilestoneAttributes,
		type UpdateIncidentMilestoneAttributes,
	} from "$lib/api";
	import { type ZonedDateTime, fromAbsolute, getLocalTimeZone, parseAbsolute } from "@internationalized/date";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import DateTimePickerField from "$components/date-time-field/DateTimePickerField.svelte";

	import { createMentionEditor } from "$components/tiptap-editor/editors";
	import { getIconForIncidentMilestoneKind, getNextOrderedMilestone, getPreviousOrderedMilestone } from "./milestones";
	import { useIncidentViewState } from "$features/incident";
	import { watch } from "runed";
	import { useIncidentTimeline } from "../timelineState.svelte";

	type Props = {
		milestone?: IncidentMilestone;
		otherMilestones: IncidentMilestone[];
		onClose: () => void;
		onSaved: (milestone: IncidentMilestone) => void;
	};
	const { milestone, otherMilestones, onClose, onSaved }: Props = $props();

	const timeline = useIncidentTimeline();
	const incidentTimeBounds = $derived(timeline.incidentWindow);

	const incidentViewState = useIncidentViewState();
	const incident = $derived(incidentViewState.incident);

	type MilestoneKindOption = {
		label: string;
		value: IncidentMilestoneAttributes["kind"];
		hint: string;
		unique?: boolean;
	};
	const milestoneKindOptions: MilestoneKindOption[] = [
		{
			label: "Impact Start",
			value: "impact",
			hint: "Impact begins",
			unique: true,
		},
		{
			label: "Detection",
			value: "detection",
			hint: "Impact is detected (monitoring, alerts, user reports, etc)",
		},
		{
			label: "Response",
			value: "investigation",
			hint: "A human is investigating",
		},
		{
			label: "Mitigation",
			value: "mitigation",
			hint: "Impact is mitigated",
		},
		{
			label: "Resolution",
			value: "resolution",
			hint: "Impact is resolved",
		},
	];
	const existingKinds = $derived(new Set(otherMilestones.map(m => m.attributes.kind)));
	const validOptions = $derived(milestoneKindOptions.filter(o => (!o.unique || !existingKinds.has(o.value))));

	type DescriptionEditor = ReturnType<typeof createMentionEditor> | null;
	type MilestoneKind = IncidentMilestoneAttributes["kind"];

	let kind = $state<MilestoneKind>(milestone?.attributes.kind ?? "impact");
	watch(() => validOptions, v => {
		const defaultOption = v.at(0);
		if (defaultOption) kind = defaultOption.value;
	});
	let descriptionEditor = $state<DescriptionEditor>(null);

	const timezone = $derived(incidentViewState.timezone);

	const incidentStart = $derived(fromAbsolute(incidentTimeBounds.start, timezone));

	const prevMs = $derived(getPreviousOrderedMilestone(kind, otherMilestones, timezone));
	const timeMin = $derived(prevMs ? parseAbsolute(prevMs.attributes.timestamp, timezone) : undefined);
	const nextMs = $derived(getNextOrderedMilestone(kind, otherMilestones, timezone));
	const timeMax = $derived(nextMs ? parseAbsolute(nextMs.attributes.timestamp, timezone) : undefined);

	const msTimestamp = milestone?.attributes.timestamp;
	const defaultTimestamp = $derived(msTimestamp ? parseAbsolute(msTimestamp, timezone) : incidentStart);
	let timestamp = $state<ZonedDateTime>();
	const timestampValue = $derived((timestamp || defaultTimestamp).toAbsoluteString());

	const saveEnabled = $derived(true);

	const mountDescriptionEditor = () => {
		const content = ""; // TODO
		descriptionEditor = createMentionEditor(content, "cursor-text focus:outline-none min-h-20");
		return () => {
			descriptionEditor?.destroy();
			descriptionEditor = null;
		};
	};
	onMount(mountDescriptionEditor);

	const getDescriptionContent = () => {
		if (!descriptionEditor) return "";
		return JSON.stringify(descriptionEditor.getJSON());
	};

	const onSuccess = ({ data: milestone }: { data: IncidentMilestone }) => onSaved(milestone);

	const createMut = createMutation(() => ({ ...createIncidentMilestoneMutation(), onSuccess }));
	const updateMut = createMutation(() => ({ ...updateIncidentMilestoneMutation(), onSuccess }));
	const loading = $derived(updateMut?.isPending || createMut?.isPending);

	const doCreate = () => {
		if (!incident || !createMut) return;
		const path = { id: $state.snapshot(incident.id) };
		const attributes: CreateIncidentMilestoneAttributes = {
			kind: $state.snapshot(kind),
			timestamp: timestampValue,
			description: getDescriptionContent(),
		};
		createMut.mutate({ path, body: { attributes } });
	};

	const doEdit = () => {
		if (!milestone || !updateMut) return;
		const path = { id: $state.snapshot(milestone.id) };
		const attributes: UpdateIncidentMilestoneAttributes = {
			kind: $state.snapshot(kind),
			timestamp: timestampValue,
			description: getDescriptionContent(),
		};
		updateMut.mutate({ path, body: { attributes } });
	};

	const onConfirm = () => {
		if (milestone) doEdit();
		else doCreate();
	};
</script>

<div class="flex flex-col min-h-0 max-h-full flex-1 gap-2 p-2">
	<DateTimePickerField 
		label="Time"
		current={timestamp || incidentStart}
		onChange={(ts) => (timestamp = ts)}
		exactTime
		rangeMin={timeMin}
		rangeMax={timeMax}
	/>

	<Field label="Kind" icon={mdiShape}>
		<ToggleGroup bind:value={kind} variant="fill" inset class="w-full">
			{#each milestoneKindOptions as opt}
				{#if !opt.unique || !existingKinds.has(opt.value)}
					<ToggleOption value={opt.value}>
						<Tooltip title={opt.hint}>
							<span class="flex items-center justify-center gap-2 px-2">
								<Icon data={getIconForIncidentMilestoneKind(opt.value)} />
								{opt.label}
							</span>
						</Tooltip>
					</ToggleOption>
				{/if}
			{/each}
		</ToggleGroup>
	</Field>

	<Field label="Description" classes={{ root: "grow", container: "h-full", input: "block" }}>
		{#if descriptionEditor}
			<TiptapEditor bind:editor={descriptionEditor} />
		{/if}
	</Field>

	<div class="flex justify-end">
		<ConfirmButtons {loading} closeText="Cancel" confirmText="Save" {saveEnabled} {onClose} {onConfirm} />
	</div>
</div>
