<script lang="ts">
	import type { OncallRoster } from "$lib/api";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { mdiChevronRight } from "@mdi/js";
	import { ListItem } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";

	type Props = {
		rosters: OncallRoster[];
	}
	let { rosters }: Props = $props();
</script>

{#each rosters as roster (roster.id)}
	<a href="/oncall/rosters/{roster.id}">
		<ListItem title={roster.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
			<svelte:fragment slot="avatar">
				<Avatar kind="roster" size={32} id={roster.id} />
			</svelte:fragment>
			<div slot="actions">
				<Icon data={mdiChevronRight} />
			</div>
		</ListItem>
	</a>
{/each}
