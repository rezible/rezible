<script lang="ts">
	import Avatar from "$components/common/entity-avatar/EntityAvatar.svelte";
	import type { OncallRoster } from "$lib/api";
	import RiArrowRightSLine from "remixicon-svelte/icons/arrow-right-s-line";
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
			<Button icon={RiArrowRightSLine} class="p-2 text-surface-content/50" />
		</div>
	</ListItem-->
</a>
