<script lang="ts">
	import { Button } from "$components/ui/button";
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
		<Button onclick={() => {view.saveEdit()}} {disabled} color="primary">Save</Button>
	{:else}
		<Button onclick={() => {view.editing = true}} {disabled} color="primary">Edit</Button>
	{/if}
</div>