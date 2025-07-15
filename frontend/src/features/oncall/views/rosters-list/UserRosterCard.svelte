<script lang="ts">
	import { Button } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { cls } from "@layerstack/tailwind";
	import { page } from "$app/state";
	import Avatar from "$components/avatar/Avatar.svelte";
	import Header from "$components/header/Header.svelte";
	import Card from "$components/card/Card.svelte";

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
		classes={{
			headerContainer: "py-3 px-4",
			root: cls(
				"w-full",
				viewing
					? "border-secondary/40 bg-secondary/10"
					: "hover:border-secondary hover:bg-secondary/25"
			),
		}}
	>
		{#snippet header()}
			<Header {title}>
				{#snippet avatar()}
					{#if icon}
						<Icon data={icon} size={32} />
					{:else if rosterId}
						<Avatar kind="roster" id={rosterId} size={32} />
					{/if}
				{/snippet}
			</Header>
		{/snippet}
	</Card>
</Button>
