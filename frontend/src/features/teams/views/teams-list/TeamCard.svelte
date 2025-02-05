<script lang="ts">
	import Avatar from "$components/avatar/Avatar.svelte";
	import { mdiDotsGrid } from "@mdi/js";
	import { Button, Card, cls, Header, Icon } from "svelte-ux";

	interface Props {
		title: string;
		teamId: string;
	}
	let { title, teamId }: Props = $props();
</script>

<Button href="/teams/{teamId}" class="p-0">
	<Card class={"hover:border-primary-600"} classes={{ headerContainer: "py-3 px-4" }}>
		<Header {title} slot="header">
			<div slot="avatar">
				{#if teamId}
					<Avatar kind="team" id={teamId} size={32} />
				{:else}
					<Icon size={32} data={mdiDotsGrid} />
				{/if}
			</div>
		</Header>
	</Card>
</Button>
