<script lang="ts">
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import {
		getNextOncallShiftOptions,
		getOncallShiftHandoverOptions,
		getOncallShiftHandoverTemplateOptions,
		sendOncallShiftHandoverMutation,
		type OncallShift,
		type OncallShiftHandover,
		type OncallShiftHandoverTemplate,
	} from "$lib/api";
	import { Button, Icon } from "svelte-ux";
	import { mdiSend, mdiPhoneForward } from "@mdi/js";
	import { getToastState } from "$features/app/lib/toasts.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import ShiftAnnotationsList from "$features/oncall/components/shift-annotations/ShiftAnnotationsList.svelte";
	import { handoverState } from "../handover.svelte";
	import ReportEditor from "./ReportEditor.svelte";
	import ShiftReviewQuestionsDialog from "./ShiftReviewQuestionsDialog.svelte";

	type Props = { shift: OncallShift; handover?: OncallShiftHandover };
	const { shift, handover }: Props = $props();

	const shiftId = $derived(shift.id);

	const templateId = $derived(shift.attributes.roster.attributes.handoverTemplateId);
	const templateQuery = createQuery(() =>
		getOncallShiftHandoverTemplateOptions({ path: { id: templateId } })
	);

	let showReviewDialog = $state(false);
</script>

<div class="px-3 flex-1 flex flex-col gap-2 max-h-full min-h-0 overflow-y-auto">
	<div class="grid grid-cols-2 gap-2 flex-1 min-h-0 max-h-full">
		<div
			class="flex flex-col gap-2 min-h-0 max-h-full overflow-y-auto border rounded-lg p-3 bg-surface-200"
		>
			<LoadingQueryWrapper query={templateQuery}>
				{#snippet view(template: OncallShiftHandoverTemplate)}
					<ReportEditor {shift} {template} {handover} />
				{/snippet}
			</LoadingQueryWrapper>
		</div>

		<div
			class="flex flex-col gap-2 min-h-0 border rounded-lg p-2 h-full max-h-full overflow-hidden"
		>
			<ShiftAnnotationsList editable={!handoverState.sent} {shiftId} />
		</div>
	</div>
</div>

{#if showReviewDialog}
	<ShiftReviewQuestionsDialog {shiftId} />
{/if}
