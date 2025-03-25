<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallShiftHandoverOptions, tryUnwrapApiError } from "$lib/api";
	import { shiftIdCtx } from "$features/oncall/lib/context.svelte";
	import { shiftState } from "$features/oncall/views/shift/shift.svelte";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";
	import ShiftHandoverEditor from "./editor/ShiftHandoverEditor.svelte";

	const shiftId = shiftIdCtx.get();
	const shift = $derived(shiftState.shift);

	const handoverQuery = createQuery(() => getOncallShiftHandoverOptions({ path: { id: shiftId } }));
	const handoverQueryError = $derived(handoverQuery.error ? tryUnwrapApiError(handoverQuery.error) : undefined);
	const handover = $derived(handoverQuery.data?.data);
	const isError = $derived(handoverQuery.isError && handoverQueryError?.status !== 404);

	const sentAt = $derived(handover && new Date(handover.attributes.sentAt));
	const isSent = $derived(sentAt && sentAt.valueOf() > 0);
</script>

{#if handoverQuery.isPending}
	<LoadingIndicator />
{:else if isError}
	<span>error: {JSON.stringify(handoverQueryError)}</span>
{:else}
	{#if isSent}
		<span>handover already sent: {handover?.id}</span>
	{:else if shift}
		<ShiftHandoverEditor {shift} {handover} />
	{/if}
{/if}