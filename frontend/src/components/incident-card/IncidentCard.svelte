<script lang="ts">
	import type { Incident } from "$lib/api";
	import { mdiChevronRight } from "@mdi/js";
	import Avatar from "$components/avatar/Avatar.svelte";
	import Icon from "$components/icon/Icon.svelte";

	type Props = {
		incident: Incident;
	};
	const { incident }: Props = $props();
</script>

<a href="/incidents/{incident.attributes.slug}">
	incident card
	<!--ListItem title={incident.attributes.title} classes={{ root: "hover:bg-secondary-900" }}>
		<svelte:fragment slot="avatar">
			<Avatar id={incident.id} square kind="incident" />
		</svelte:fragment>
		<div slot="actions">
			<Icon data={mdiChevronRight} size={24} classes={{root: "text-surface-content/50"}} />
		</div>
	</ListItem-->
</a>