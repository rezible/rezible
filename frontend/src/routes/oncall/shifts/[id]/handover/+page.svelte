<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { page } from "$app/state";
	import {
		getOncallShiftHandoverOptions,
		getOncallShiftOptions,
		tryUnwrapApiError,
		type OncallShift,
	} from "$lib/api";
	import ShiftHandoverEditor from "$features/oncall/views/shift-handover-editor/ShiftHandoverEditor.svelte";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";

	const shiftId = $derived(page.params.id);
	const shiftQuery = createQuery(() => getOncallShiftOptions({ path: { id: shiftId } }));

	const handoverQuery = createQuery(() => getOncallShiftHandoverOptions({ path: { id: shiftId } }));
	const handover = $derived(handoverQuery.data?.data);

	const handoverQueryError = $derived(handoverQuery.error ? tryUnwrapApiError(handoverQuery.error) : null);
	const isSent = $derived(handover && handover.attributes.sent_at.valueOf() > 0);
</script>

{#key shiftId}
	<LoadingQueryWrapper query={shiftQuery}>
		{#snippet view(shift: OncallShift)}
			{#if handover || handoverQueryError?.status === 404}
				{#if !isSent}
					<ShiftHandoverEditor {shift} {handover} />
				{:else}
					<span>show handover: {handover?.id}</span>
				{/if}
			{/if}
		{/snippet}
	</LoadingQueryWrapper>
{/key}
