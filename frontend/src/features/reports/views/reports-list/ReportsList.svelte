<script lang="ts">
	import { mdiFilter, mdiPlus } from "@mdi/js";
	import { Button, Header, Collapse, Icon } from "svelte-ux";

</script>

<div class="flex flex-col gap-2">
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
</div>