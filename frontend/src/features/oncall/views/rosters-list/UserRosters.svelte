<script lang="ts">
	import { type OncallRoster } from '$lib/api';
    import UserRosterCard from './UserRosterCard.svelte';

	type Props = {
		rosters: OncallRoster[];
	}
	const { rosters }: Props = $props();
</script>

<div class="flex flex-row items-end gap-4">
	<div class="flex flex-col h-full">
		<div class="h-fit">
			<span>Your Rosters</span>
		</div>
		<div class="flex-1 flex items-end gap-4">
			{#each rosters as r (r.id)}
				<UserRosterCard
					title={r.attributes.name}
					href="/oncall/rosters/{r.attributes.slug}"
					rosterId={r.id}
				/>
			{/each}
		</div>
	</div>
</div>