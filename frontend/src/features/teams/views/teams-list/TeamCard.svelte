<script lang="ts">
	import Avatar from "$components/avatar/Avatar.svelte";
	import Header from "$src/components/header/Header.svelte";
	import { mdiDotsGrid } from "@mdi/js";
	import { Button, Card, Icon } from "svelte-ux";

	type Props = {
		title: string;
		teamId: string;
	}
	let { title, teamId }: Props = $props();
</script>

<Button href="/teams/{teamId}" class="p-0">
	<Card class={"hover:border-primary-600"} classes={{ headerContainer: "py-3 px-4" }}>
		<Header {title} slot="header">
			{#snippet avatar()}
				{#if teamId}
					<Avatar kind="team" id={teamId} size={32} />
				{:else}
					<Icon size={32} data={mdiDotsGrid} />
				{/if}
			{/snippet}
		</Header>
	</Card>
</Button>
