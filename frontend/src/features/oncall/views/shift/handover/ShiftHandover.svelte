<script lang="ts">
	import { createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { getOncallShiftHandoverOptions } from "$lib/api";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";

	import { ShiftHandoverContent, ShiftHandoverEditorState } from "$features/oncall/components/shift-handover-content";

	import { shiftViewStateCtx } from "../context.svelte";

	import AnnotatedEventsList from "./AnnotatedEventsList.svelte";
	import ShiftReviewQuestionsDialog from "./ShiftReviewQuestionsDialog.svelte";
	import SendHandoverButton from "./SendHandoverButton.svelte";

	const viewState = shiftViewStateCtx.get();
	const shiftId = $derived(viewState.shiftId);

	const handoverQueryOpts = $derived(getOncallShiftHandoverOptions({ path: { id: shiftId } }));
	const handoverQuery = createQuery(() => handoverQueryOpts);
	const handover = $derived(handoverQuery.data?.data);

	const allowEditing = true;
	const handoverEditorState = new ShiftHandoverEditorState(() => handover, allowEditing);

	const queryClient = useQueryClient();
	const onUpdated = () => queryClient.invalidateQueries(handoverQueryOpts);

	let showReviewDialog = $state(false);

	const loading = $derived(handoverQuery.isLoading);
</script>

<div class="flex-1 flex flex-col gap-2 max-h-full h-full min-h-0 overflow-auto">
	<div class="flex-1 flex gap-2 max-h-full min-h-0 overflow-y-auto">
		{#if handover && !handoverEditorState.isSent}
			<div class="w-1/3 flex flex-col gap-2 min-h-0 h-full max-h-full overflow-hidden">
				<AnnotatedEventsList {handover} {onUpdated} />
			</div>
		{/if}

		<div class="flex-1 flex flex-col justify-between gap-2 h-full max-h-full">
			{#if loading}
				<LoadingIndicator />
			{:else if handover}
				<ShiftHandoverContent handoverState={handoverEditorState} />
			{/if}
		</div>
	</div>

	{#if handoverEditorState.editable}
		<div class="border-t pt-2 flex items-center justify-end">
			<SendHandoverButton handoverState={handoverEditorState} />
		</div>
	{/if}
</div>

{#if showReviewDialog}
	<ShiftReviewQuestionsDialog />
{/if}