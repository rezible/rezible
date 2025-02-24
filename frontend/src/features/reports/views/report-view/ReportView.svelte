<script lang="ts">
	import { Button, ButtonGroup, Header, Icon } from "svelte-ux";
	import IncidentsGraph from "$features/reports/components/incidents-graph/IncidentsGraph.svelte";
	import { mdiContentDuplicate, mdiCopyleft, mdiPlus, mdiStar, mdiStarOutline, mdiTrashCan } from "@mdi/js";
	import { setPageBreadcrumbs } from "$lib/appState.svelte";

	type Props = {id: string;}
	const { id }: Props = $props();

	let starred = $state(true);

	const report = {id: "foo", attributes: {title: "Test Report", author: "tex", slug: "test"}};

	setPageBreadcrumbs(() => [
		{ label: "Reports", href: "/reports" },
		{ label: report.attributes.title },
	]);
</script>

<div class="flex flex-col gap-2">
	<Header title={report.attributes.title} subheading="" classes={{ root: "h-11" }}>
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

	<IncidentsGraph />

	<Button variant="fill-light">
		<span class="flex gap-2 items-center">
			Add Cell
			<Icon data={mdiPlus} />
		</span>
	</Button>
</div>