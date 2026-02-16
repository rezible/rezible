<script lang="ts">
	import Avatar from "$components/avatar/Avatar.svelte";
	import type { OncallRoster } from "$lib/api";
	import { mdiChevronRight } from "@mdi/js";
	import { Button } from "$components/ui/button";

	type Props = {
		roster: OncallRoster;
	};
	const { roster }: Props = $props();
</script>

<a href="/rosters/{roster.attributes.slug}">
	<!--ListItem title={roster.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
		<svelte:fragment slot="avatar">
			<Avatar kind="roster" size={32} id={roster.id} />
		</svelte:fragment>
		<div slot="actions">
			<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
		</div>
	</ListItem-->
</a>