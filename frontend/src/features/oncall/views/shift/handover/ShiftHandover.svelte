<script lang="ts">
	import { createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { getOncallShiftHandoverOptions } from "$lib/api";
	import { shiftIdCtx } from "$features/oncall/lib/context.svelte";
	import { shiftState } from "$features/oncall/views/shift/shift.svelte";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";
	import ShiftHandoverContent from "$features/oncall/components/shift-handover-content/ShiftHandoverContent.svelte";
	import AnnotatedEventsList from "./AnnotatedEventsList.svelte";
	import ShiftReviewQuestionsDialog from "./ShiftReviewQuestionsDialog.svelte";

	const shiftId = shiftIdCtx.get();
	const shift = $derived(shiftState.shift);

	const queryOpts = $derived(getOncallShiftHandoverOptions({ path: { id: shiftId } }));
	const handoverQuery = createQuery(() => queryOpts);
	const handover = $derived(handoverQuery.data?.data);

	const queryClient = useQueryClient();
	const onUpdated = () => queryClient.invalidateQueries(queryOpts);

	const isSent = $derived(handover && new Date(handover.attributes.sentAt).valueOf() > 0);

	let showReviewDialog = $state(false);

	const loading = $derived(handoverQuery.isLoading);
</script>

<div class="flex-1 flex gap-2 max-h-full h-full min-h-0 overflow-y-auto">
	{#if shift && handover && !isSent}
		<div class="w-1/3 flex flex-col gap-2 min-h-0 border rounded-lg p-2 h-full max-h-full overflow-hidden">
			<AnnotatedEventsList {shift} {handover} {onUpdated} />
		</div>
	{/if}

	<div class="flex-1 flex flex-col gap-2 h-full max-h-full">
		{#if loading}
			<LoadingIndicator />
		{:else if handover}
			<ShiftHandoverContent {shiftId} {handover} editable />
		{/if}
	</div>
</div>

{#if showReviewDialog}
	<ShiftReviewQuestionsDialog {shiftId} />
{/if}