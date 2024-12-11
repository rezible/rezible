<script lang="ts">
	import type { OncallRoster } from "$lib/api";
	import { AvatarSunset } from "svelte-boring-avatars";

	interface Props { rosters: OncallRoster[] };
	let { rosters }: Props = $props();
</script>

<div class="flex flex-col gap-2 w-fit">
	{#each rosters as roster (roster.id)}
		<a href="/oncall/rosters/{roster.id}">
			<div class="flex items-center gap-2 px-2 hover:bg-surface-200 p-1 rounded-lg">
				<AvatarSunset name={roster.id} size={32} />
				<span>{roster.attributes.name}</span>
			</div>
		</a>
	{/each}
</div>