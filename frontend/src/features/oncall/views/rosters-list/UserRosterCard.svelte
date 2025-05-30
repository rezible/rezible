<script lang="ts">
	import { Button, Card } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { cls } from '@layerstack/tailwind';
	import { page } from "$app/state";
	import Avatar from "$components/avatar/Avatar.svelte";
	import Header from "$src/components/header/Header.svelte";

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
			viewing ? "border-secondary/40 bg-secondary/10" : "hover:border-secondary hover:bg-secondary/25"
		)}
		classes={{ headerContainer: "py-3 px-4", root: "w-full" }}
	>
		<Header {title} slot="header">
			{#snippet avatar()}
				{#if icon}
					<Icon data={icon} size={32} />
				{:else if rosterId}
					<Avatar kind="roster" id={rosterId} size={32} />
				{/if}
			{/snippet}
		</Header>
	</Card>
</Button>
