<script lang="ts">
	import { Field, Icon, Tooltip, ToggleGroup, ToggleOption } from "svelte-ux";
	import { mdiShape } from "@mdi/js";
	import { EditorContent } from "svelte-tiptap";
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
	import { type ZonedDateTime, parseAbsoluteToLocal } from "@internationalized/date";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import DateTimePickerField from "$components/date-time-field/DateTimePickerField.svelte";
	import { incidentCtx } from "$features/incidents/lib/context";
	import { createMentionEditor } from "$features/incidents/lib/editor.svelte";
	import { getIconForIncidentMilestoneKind, getNextOrderedMilestone, getPreviousOrderedMilestone } from "$src/features/incidents/lib/milestones";

	type Props = {
		milestone?: IncidentMilestone;
		otherMilestones: IncidentMilestone[];
		onClose: () => void;
		onSaved: (milestone: IncidentMilestone) => void;
	};
	const { milestone, otherMilestones, onClose, onSaved }: Props = $props();

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

	type DescriptionEditor = ReturnType<typeof createMentionEditor> | null;
	type MilestoneKind = IncidentMilestoneAttributes["kind"];

	const incident = incidentCtx.get();

	let kind = $state<MilestoneKind>(milestone?.attributes.kind ?? "impact");
	let descriptionEditor = $state<DescriptionEditor>(null);

	const prevMs = $derived(getPreviousOrderedMilestone(kind, otherMilestones));
	const timeMin = $derived(prevMs ? parseAbsoluteToLocal(prevMs.attributes.timestamp) : undefined);
	const nextMs = $derived(getNextOrderedMilestone(kind, otherMilestones));
	const timeMax = $derived(nextMs ? parseAbsoluteToLocal(nextMs.attributes.timestamp) : undefined);

	const defaultTimestamp = milestone?.attributes.timestamp ?? incident.attributes.openedAt;
	let timestamp = $state<ZonedDateTime>(parseAbsoluteToLocal(defaultTimestamp));

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
			timestamp: timestamp.toAbsoluteString(),
			description: getDescriptionContent(),
		};
		createMut.mutate({ path, body: { attributes } });
	};

	const doEdit = () => {
		if (!milestone || !updateMut) return;
		const path = { id: $state.snapshot(milestone.id) };
		const attributes: UpdateIncidentMilestoneAttributes = {
			kind: $state.snapshot(kind),
			timestamp: timestamp.toAbsoluteString(),
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
	<Field label="Kind" icon={mdiShape}>
		<ToggleGroup bind:value={kind} variant="fill" inset class="w-full">
			{#each milestoneKindOptions as opt}
				<ToggleOption value={opt.value}>
					<Tooltip title={opt.hint}>
						<span class="flex items-center justify-center gap-2 px-2">
							<Icon data={getIconForIncidentMilestoneKind(opt.value)} />
							{opt.label}
						</span>
					</Tooltip>
				</ToggleOption>
			{/each}
		</ToggleGroup>
	</Field>

	<DateTimePickerField label="Time" current={timestamp} onChange={(ts) => (timestamp = ts)} exactTime rangeMin={timeMin} rangeMax={timeMax} />

	<Field label="Description" classes={{ root: "grow", container: "h-full", input: "block" }}>
		{#if descriptionEditor}
			<EditorContent editor={descriptionEditor} />
		{/if}
	</Field>

	<div class="flex justify-end">
		<ConfirmButtons {loading} closeText="Cancel" confirmText="Save" {saveEnabled} {onClose} {onConfirm} />
	</div>
</div>
