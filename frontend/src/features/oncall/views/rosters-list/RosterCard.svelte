<script lang="ts">
	import type { OncallRoster } from '$lib/api';
	import { mdiStar, mdiStarOutline, mdiArrowRight } from '@mdi/js';
	import { Card, Header, Button, Icon } from 'svelte-ux';

	interface Props {
		roster: OncallRoster;
		pinned?: boolean;
	};
	let { 
		roster, 
		pinned = false,
	}: Props = $props();
</script>

<Card class="">
	<div slot="header">
		<Header title={roster.attributes.name}></Header>
	</div>

	<div slot="contents" class="flex flex-col">
		<div class="flex flex-col">
			<span></span>
			<div class="flex items-center"></div>
		</div>
	</div>

	<div slot="actions" class="px-1 flex flex-row-reverse gap-2">
		<Button href="/oncall/rosters/{roster.id}">
			View
			<Icon data={mdiArrowRight} />
		</Button>
		<Button iconOnly icon={pinned ? mdiStar : mdiStarOutline} />
	</div>
</Card>
