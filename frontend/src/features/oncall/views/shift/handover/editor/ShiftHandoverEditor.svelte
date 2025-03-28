<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import {
		getOncallShiftHandoverTemplateOptions,
		listOncallShiftAnnotationsOptions,
		listOncallShiftIncidentsOptions,
		type OncallShift,
		type OncallShiftHandover,
	} from "$lib/api";
	import ShiftAnnotationsList from "$features/oncall/components/shift-annotations/ShiftAnnotationsList.svelte";
	import { handoverState } from "../handover.svelte";
	import ReportEditor from "./ReportEditor.svelte";
	import ShiftReviewQuestionsDialog from "./ShiftReviewQuestionsDialog.svelte";
	import SendButton from "./SendButton.svelte";
	import LoadingIndicator from "$src/components/loader/LoadingIndicator.svelte";
	import { Header } from "svelte-ux";

	type Props = { shift: OncallShift; handover?: OncallShiftHandover };
	const { shift, handover }: Props = $props();

	const shiftId = $derived(shift.id);
	const sentAt = $derived(handover && new Date(handover.attributes.sentAt));
	const isSent = $derived(sentAt && sentAt.valueOf() > 0);

	let showReviewDialog = $state(false);

	const templateId = $derived(shift.attributes.roster.attributes.handoverTemplateId);
	const templateQuery = createQuery(() => ({
		...getOncallShiftHandoverTemplateOptions({ path: { id: templateId } }),
		enabled: !isSent,
	}));
	const template = $derived(templateQuery.data?.data);

	const annotationsQuery = createQuery(() => ({
		...listOncallShiftAnnotationsOptions({ path: { id: shift.id } }),
		enabled: !isSent,
	}));
	const annotations = $derived(annotationsQuery.data?.data ?? []);
	const pinnedAnnotations = $derived(annotations.filter((ann) => ann.attributes.pinned));

	const incidentsSectionPresent = $derived(handoverState.sections.some((s) => s.kind === "incidents"));
	const incidentsQuery = createQuery(() => ({
		...listOncallShiftIncidentsOptions({ path: { id: shift.id } }),
		enabled: (!isSent && incidentsSectionPresent),
	}));
	const incidents = $derived(incidentsQuery.data?.data ?? []);

	const loading = $derived(isSent ? false : templateQuery.isLoading);
</script>

<div class="flex-1 flex gap-2 max-h-full h-full min-h-0 overflow-y-auto">
	{#if !isSent}
		<div class="w-1/3 flex flex-col gap-2 min-h-0 border rounded-lg p-2 h-full max-h-full overflow-hidden">
			<ShiftAnnotationsList editable={!handoverState.sent} {shiftId} />
		</div>
	{/if}

	<div class="flex-1 flex flex-col gap-2 h-full max-h-full">
		{#if loading}
			<LoadingIndicator />
		{:else}
			<ReportEditor {template} {handover} {incidents} {pinnedAnnotations} />

			{#if !isSent}
				<div class="w-full flex justify-end px-2">
					<SendButton {shiftId} />
				</div>
			{/if}
		{/if}
	</div>
</div>

{#if showReviewDialog}
	<ShiftReviewQuestionsDialog {shiftId} />
{/if}
