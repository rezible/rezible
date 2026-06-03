<script lang="ts">
	import type { AvailableIntegration } from "$lib/api";
	import AvailableIntegrationCard from "./AvailableIntegrationCard.svelte";

	type Props = {
		provider: string;
		integrations: AvailableIntegration[];
	};
	const { provider, integrations }: Props = $props();
</script>

<div class="border p-2">
	<span>provider: {provider}</span>

	{#each integrations as integration}
		<AvailableIntegrationCard {integration} />
	{/each}
</div>