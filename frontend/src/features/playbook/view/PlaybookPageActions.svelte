<script lang="ts">
	import Button from "$components/button/Button.svelte";
	import type { PlaybookViewState } from "../lib/viewState.svelte";

	type Props = {
		view: PlaybookViewState;
	};
	const { view }: Props = $props();

	const disabled = $derived(view.loading);
</script>

<div class="self-end">
	{#if view.editing}
		<Button onclick={() => {view.cancelEditing()}} {disabled}>Cancel</Button>
		<Button onclick={() => {view.saveEdit()}} {disabled} loading={view.loading} variant="fill" color="primary">Save</Button>
	{:else}
		<Button onclick={() => {view.editing = true}} {disabled} variant="fill" color="primary">Edit</Button>
	{/if}
</div>