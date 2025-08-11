<script lang="ts">
	import { mdiChevronRight } from "@mdi/js";
	import { ListItem } from "svelte-ux";
	import Button from "$components/button/Button.svelte";
	import Header from "$components/header/Header.svelte";

	const reports = [
		{id: "foo", attributes: {title: "Test Report", author: "tex", slug: "test"}},
	];
</script>

<Header title="Starred" subheading="" classes={{ title: "text-2xl", root: "h-11" }} />

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