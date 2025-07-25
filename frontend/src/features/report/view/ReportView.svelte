<script lang="ts">
	import { Button, ButtonGroup } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiContentDuplicate, mdiPlus, mdiStar, mdiStarOutline, mdiTrashCan } from "@mdi/js";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import IncidentsGraph from "./IncidentsGraph.svelte";
	import Header from "$components/header/Header.svelte";

	type Props = {id: string;}
	const { id }: Props = $props();

	let starred = $state(true);

	const report = {id: "foo", attributes: {title: "Test Report", author: "tex", slug: "test"}};

	appShell.setPageBreadcrumbs(() => [
		{ label: "Reports", href: "/reports" },
		{ label: report.attributes.title },
	]);
</script>

<div class="flex flex-col gap-2 overflow-y-auto">
	<Header title={report.attributes.title} subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		{#snippet actions()}
			<ButtonGroup variant="fill-light">
				<Button icon={starred ? mdiStar : mdiStarOutline} on:click={() => {starred = !starred}}>
					{starred ? "Unstar" : "Star"}
				</Button>
				<Button icon={mdiContentDuplicate}>Duplicate</Button>
				<Button icon={mdiTrashCan}>Delete</Button>
			</ButtonGroup>
		{/snippet}
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