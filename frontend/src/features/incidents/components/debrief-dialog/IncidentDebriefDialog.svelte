<script lang="ts">
	import { mdiClose } from "@mdi/js";
	import type { IncidentDebrief } from "$lib/api";
	import { Button } from "$components/ui/button";
	import IncidentDebriefView from "./IncidentDebriefView.svelte";
	import Header from "$components/header/Header.svelte";

	type Props = {
		debrief: IncidentDebrief;
		open: boolean;
	};
	let { debrief, open = $bindable() }: Props = $props();
</script>

<!-- <Dialog
	bind:open
	persistent
	portal
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-7xl my-2",
		root: "p-2",
	}}
>
	<div slot="header" class="border-b p-2" let:close>
		<Header title="Debrief">
			{#snippet actions()}
				<Button onclick={() => close({ force: true })} iconOnly icon={mdiClose} />
			{/snippet}
		</Header>
	</div>

	<svelte:fragment slot="default">
		{#if debrief && open}
			<IncidentDebriefView {debrief} />
		{/if}
	</svelte:fragment>
</Dialog> -->
