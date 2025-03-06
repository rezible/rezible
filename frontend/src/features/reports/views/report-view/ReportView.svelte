<script lang="ts">
	import { Button, ButtonGroup, Header, Icon } from "svelte-ux";
	import { mdiContentDuplicate, mdiPlus, mdiStar, mdiStarOutline, mdiTrashCan } from "@mdi/js";
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import IncidentsGraph from "./IncidentsGraph.svelte";

	type Props = {id: string;}
	const { id }: Props = $props();

	let starred = $state(true);

	const report = {id: "foo", attributes: {title: "Test Report", author: "tex", slug: "test"}};

	setPageBreadcrumbs(() => [
		{ label: "Reports", href: "/reports" },
		{ label: report.attributes.title },
	]);
</script>

<div class="flex flex-col gap-2 overflow-y-auto">
	<Header title={report.attributes.title} subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		<svelte:fragment slot="actions">
			<ButtonGroup variant="fill-light">
				<Button icon={starred ? mdiStar : mdiStarOutline} on:click={() => {starred = !starred}}>
					{starred ? "Unstar" : "Star"}
				</Button>
				<Button icon={mdiContentDuplicate}>Duplicate</Button>
				<Button icon={mdiTrashCan}>Delete</Button>
			</ButtonGroup>
		</svelte:fragment>
	</Header>

	<div class="flex-1 flex flex-col gap-2 max-h-full overflow-y-auto">
		<IncidentsGraph />

		<Button variant="fill-light">
			<span class="flex gap-2 items-center">
				Add Cell
				<Icon data={mdiPlus} />
			</span>
		</Button>
	</div>
</div>