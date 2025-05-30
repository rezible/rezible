<script lang="ts">
	import { mdiDotsGrid } from "@mdi/js";
	import { Button } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import Header from "$components/header/Header.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import Card from "$components/card/Card.svelte";

	type Props = {
		title: string;
		teamId: string;
	}
	let { title, teamId }: Props = $props();
</script>

<Button href="/teams/{teamId}" class="p-0">
	<Card classes={{ root: "hover:border-primary-600", headerContainer: "py-3 px-4" }}>
		{#snippet header()}
			<Header {title}>
				{#snippet avatar()}
					{#if teamId}
						<Avatar kind="team" id={teamId} size={32} />
					{:else}
						<Icon size={32} data={mdiDotsGrid} />
					{/if}
				{/snippet}
			</Header>
		{/snippet}
	</Card>
</Button>
