<script lang="ts">
	import { mdiChevronRight, mdiFilter, mdiPlus } from "@mdi/js";
	import { Button, Header, Collapse, Icon, ListItem } from "svelte-ux";

	const reports = [
		{id: "foo", attributes: {title: "Test Report", author: "tex", slug: "test"}},
	];
</script>

<div class="flex flex-col gap-2 h-full">
	<Header title="Browse" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		<svelte:fragment slot="actions">
			<Button color="primary" variant="fill" href="/reports/new">
				<span class="flex gap-2 items-center">
					Create New
					<Icon data={mdiPlus} />
				</span>
			</Button>
		</svelte:fragment>
	</Header>

	<Collapse>
		<div slot="trigger" class="flex-1">
			<Button icon={mdiFilter} iconOnly />
			Filters
		</div>
		<div class="px-3 pb-3 border-t">
			show filters
		</div>
	</Collapse>

	<div class="flex flex-col gap-2 flex-1 overflow-y-auto">
		{#each reports as r}
			<a href="/reports/{r.attributes.slug}">
				<ListItem title={r.attributes.title} classes={{ root: "hover:bg-secondary-900" }}>
					<div slot="actions">
						<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
					</div>
				</ListItem>
			</a>
		{/each}
	</div>
</div>