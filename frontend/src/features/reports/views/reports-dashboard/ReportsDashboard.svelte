<script lang="ts">
	import { mdiChevronRight, mdiFilter, mdiChevronDown } from "@mdi/js";
	import { Button, Header, Collapse, Icon, ListItem } from "svelte-ux";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import ReportsPageActions from "$features/reports/components/reports-page-actions/ReportsPageActions.svelte";

	appShell.setPageActions(ReportsPageActions, true);

	const reports = [
		{id: "foo", attributes: {title: "Test Report", author: "tex", slug: "test"}},
	];
</script>

<div class="flex flex-col gap-2 h-full">
	<Header title="Browse" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		<div slot="actions" class="">
			<Button icon={mdiFilter} iconOnly>
				<span class="flex gap-1 items-center">
					Filters
					<Icon data={mdiChevronDown} />
				</span>
			</Button>
		</div>
	</Header>

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