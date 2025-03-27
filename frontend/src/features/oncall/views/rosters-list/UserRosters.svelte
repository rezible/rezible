<script lang="ts">
	import { type OncallRoster } from "$lib/api";
	import UserRosterCard from "./UserRosterCard.svelte";

	type Props = {
		rosters: OncallRoster[];
	};
	const { rosters }: Props = $props();
</script>

<div class="flex flex-col h-full">
	{#each rosters as r (r.id)}
		<UserRosterCard
			title={r.attributes.name}
			href="/oncall/rosters/{r.attributes.slug}"
			rosterId={r.id}
		/>
	{/each}
</div>