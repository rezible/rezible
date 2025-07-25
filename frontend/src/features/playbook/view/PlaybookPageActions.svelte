<script lang="ts">
	import { Button } from "svelte-ux";
	import type { PlaybookViewState } from "./viewState.svelte";

	type Props = {
		viewState: PlaybookViewState;
	};
	const { viewState }: Props = $props();

	const disabled = $derived(viewState.loading);
</script>

<div class="self-end">
	{#if viewState.editing}
		<Button on:click={() => {viewState.cancelEditing()}} {disabled}>Cancel</Button>
		<Button on:click={() => {viewState.saveEdit()}} {disabled} loading={viewState.loading} variant="fill" color="primary">Save</Button>
	{:else}
		<Button on:click={() => {viewState.editing = true}} {disabled} variant="fill" color="primary">Edit</Button>
	{/if}
</div>