<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import {
		getOncallShiftHandoverOptions,
		tryUnwrapApiError,
		type OncallShift,
	} from "$lib/api";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import ShiftHandoverEditor from "./handover-editor/ShiftHandoverEditor.svelte";
	import PageActions from "./PageActions.svelte";

	type Props = {
		shift: OncallShift;
	};
	const { shift }: Props = $props();

	appShell.setPageActions(PageActions, false);

	const handoverQuery = createQuery(() => getOncallShiftHandoverOptions({ path: { id: shift.id } }));
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
	{:else}
		<ShiftHandoverEditor {shift} {handover} />
	{/if}
{/if}