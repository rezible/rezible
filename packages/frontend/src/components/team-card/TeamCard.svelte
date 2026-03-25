<script lang="ts">
	import type { Team } from "$lib/api";
	import { mdiChevronRight } from "@mdi/js";
	import Avatar from "$components/avatar/Avatar.svelte";
	import Icon from "$components/icon/Icon.svelte";

	type Props = {
		team: Team;
	};
	const { team }: Props = $props();
</script>

<a href="/teams/{team.attributes.slug}">
	<span>team card</span>
	<!--ListItem title={team.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
		<svelte:fragment slot="avatar">
			<Avatar kind="team" size={32} id={team.id} />
		</svelte:fragment>
		<div slot="actions">
			<Icon data={mdiChevronRight} size={24} classes={{root: "text-surface-content/50"}} />
		</div>
	</ListItem-->
</a>