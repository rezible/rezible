<script lang="ts">
	import { Button, Card, cls, Header, Icon } from "svelte-ux";
	import { page } from "$app/state";
	import Avatar from "$components/avatar/Avatar.svelte";

	type Props = {
		title: string;
		href: string;
		icon?: string;
		rosterId?: string;
	};
	const { title, href, icon, rosterId }: Props = $props();

	const viewing = $derived(href == page.url.pathname);
</script>

<Button {href} class={"p-0 hover:bg-none text-left"} on:click={() => {}}>
	<Card
		class={cls(
			viewing
				? "border-secondary/40 bg-secondary/10"
				: "hover:border-secondary hover:bg-secondary/25"
		)}
		classes={{ headerContainer: "py-3 px-4" }}
	>
		<Header {title} slot="header">
			<div slot="avatar">
				{#if icon}
					<Icon data={icon} size={32} />
				{:else if rosterId}
					<Avatar kind="roster" id={rosterId} size={32} />
				{/if}
			</div>
		</Header>
	</Card>
</Button>
