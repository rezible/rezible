<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import {
		getOncallShiftHandoverOptions,
		tryUnwrapApiError,
		type OncallShift,
	} from "$lib/api";
	import ShiftHandoverEditor from "./handover-editor/ShiftHandoverEditor.svelte";

	type Props = {
		shift: OncallShift;
	};
	const { shift }: Props = $props();

	const handoverQuery = createQuery(() => getOncallShiftHandoverOptions({ path: { id: shift.id } }));
	const handover = $derived(handoverQuery.data?.data);

	const handoverQueryError = $derived(handoverQuery.error ? tryUnwrapApiError(handoverQuery.error) : undefined);
	const sentAt = $derived(handover && new Date(handover.attributes.sentAt));
	const isSent = $derived(sentAt && sentAt.valueOf() > 0);
</script>

{#if handover || handoverQueryError?.status === 404}
	{#if isSent}
		<span>show sent handover: {handover?.id}</span>
	{:else}
		<ShiftHandoverEditor {shift} {handover} />
	{/if}
{/if}