<script lang="ts">
	import Button from "$components/button/Button.svelte";
	import { mdiDotsVertical } from "@mdi/js";
	import Header from "$components/header/Header.svelte";

</script>

<div class="h-fit flex flex-col gap-2 border rounded px-4 py-2">
	<Header>
		{#snippet title()}
			<span class="text-lg">
				Incidents <span class="opacity-80">(past 30 days)</span>
			</span>
		{/snippet}
		{#snippet actions()}
			<Button icon={mdiDotsVertical} iconOnly />
		{/snippet}
	</Header>

	<div class="h-[300px] p-4 border rounded bg-surface-100">
		
	</div>
</div>
