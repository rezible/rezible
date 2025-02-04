<script lang="ts">
	import type { OncallRoster } from "$lib/api";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { mdiChevronRight } from "@mdi/js";
	import { Icon, ListItem } from "svelte-ux";

	interface Props {
		rosters: OncallRoster[];
	}
	let { rosters }: Props = $props();
</script>

{#each rosters as roster (roster.id)}
	<a href="/oncall/rosters/{roster.id}">
		<ListItem
			title={roster.attributes.name}
			classes={{ root: "hover:bg-secondary-900" }}
		>
			<svelte:fragment slot="avatar">
				<Avatar kind="roster" size={32} id={roster.id} />
			</svelte:fragment>
			<div slot="actions">
				<Icon data={mdiChevronRight} />
			</div>
		</ListItem>
		<div class="h-20"></div>
	</a>
{/each}
