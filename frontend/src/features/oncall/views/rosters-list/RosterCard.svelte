<script lang="ts">
	import type { OncallRoster } from "$lib/api";
	import Header from "$components/header/Header.svelte";
	import { mdiStar, mdiStarOutline, mdiArrowRight } from "@mdi/js";
	import { Button } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import Card from "$components/card/Card.svelte";

	type Props = {
		roster: OncallRoster;
		pinned?: boolean;
	};
	let { roster, pinned = false }: Props = $props();
</script>

<Card>
	{#snippet header()}
		<Header title={roster.attributes.name}></Header>
	{/snippet}

	{#snippet contents()}
		<div class="flex flex-col">
			<div class="flex flex-col">
				<span></span>
				<div class="flex items-center"></div>
			</div>
		</div>
	{/snippet}

	{#snippet actions()}
		<div class="px-1 flex flex-row-reverse gap-2">
			<Button href="/oncall/rosters/{roster.id}">
				View
				<Icon data={mdiArrowRight} />
			</Button>
			<Button iconOnly icon={pinned ? mdiStar : mdiStarOutline} />
		</div>
	{/snippet}
</Card>
