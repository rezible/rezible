<script lang="ts">
	import { mdiChevronRight, mdiFilter, mdiChevronDown } from "@mdi/js";
	import { Button, ListItem } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import ReportsListPageActions from "./ReportsListPageActions.svelte";
	import Header from "$components/header/Header.svelte";

	appShell.setPageActions(ReportsListPageActions, false);

	const reports = [
		{id: "foo", attributes: {title: "Test Report", author: "tex", slug: "test"}},
	];
</script>

<div class="flex flex-col gap-2 h-full">
	<Header title="Browse" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		{#snippet actions()}
			<Button icon={mdiFilter} iconOnly>
				<span class="flex gap-1 items-center">
					Filters
					<Icon data={mdiChevronDown} />
				</span>
			</Button>
		{/snippet}
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