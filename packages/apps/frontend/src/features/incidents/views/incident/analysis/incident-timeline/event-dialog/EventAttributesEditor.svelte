<script lang="ts">
	import type { Component } from "svelte";
	
	import RiNodeTree from "remixicon-svelte/icons/node-tree";
	import RiStackLine from "remixicon-svelte/icons/stack-line";
	import RiLink from "remixicon-svelte/icons/link";
	import RiFlowChart from "remixicon-svelte/icons/flow-chart";

	import Header from "$src/components/layout/header/Header.svelte";

	import EventDetailsPanel from "./attribute-panels/EventDetails.svelte";
	import DecisionContextPanel from "./attribute-panels/DecisionContext.svelte";
	import ContributingFactorsPanel from "./attribute-panels/ContributingFactors.svelte";
	import EvidencePanel from "./attribute-panels/Evidence.svelte";
	import SystemContextPanel from "./attribute-panels/SystemContext.svelte";
	import { useEventDialog } from "./controller.svelte";

	const eventDialog = useEventDialog();
</script>

<div class="flex flex-row min-h-0 max-h-full flex-1 gap-2 p-2">
	<div class="flex flex-col gap-2 pl-1">
		<Header title="Details" />

		<div class="flex flex-col flex-1 gap-2 overflow-y-auto">
			<EventDetailsPanel />
		</div>
	</div>

	<div class="flex flex-col gap-2 overflow-y-auto flex-1">
		<Header title="Context" />

		<div class="flex-1 flex flex-col gap-2 overflow-y-auto pr-1">
			{#if eventDialog.attributes.kind === "decision"}
				{@render componentTraitPanel(
					"Decision Context",
					"Document the options, constraints, and reasoning behind this choice",
					RiNodeTree,
					DecisionContextPanel
				)}
			{/if}

			{@render componentTraitPanel(
				"Contributing Factors",
				"Identify pressures and conditions that shaped this event",
				RiStackLine,
				ContributingFactorsPanel
			)}

			{@render componentTraitPanel(
				"Evidence & Links",
				"Add links to logs, metrics, discussions, and other supporting information",
				RiLink,
				EvidencePanel
			)}

			{@render componentTraitPanel(
				"System Context",
				"Document the relevant system components and their conditions at this time",
				RiFlowChart,
				SystemContextPanel
			)}
		</div>
	</div>
</div>

{#snippet componentTraitPanel(title: string, subheading: string, icon: Component, PanelComponent: Component)}
	<div class="p-2 border rounded">
		<!-- <Collapse open classes={{ root: "overflow-x-hidden", content: "p-2" }}>
		<ListItem
			slot="trigger"
			{title}
			{subheading}
			{icon}
			classes={{ root: "pl-0" }}
			avatar={{
				class: "bg-foreground/50 text-card/90",
			}}
			class="flex-1"
			noShadow
		/>
		<PanelComponent></PanelComponent>
	</Collapse> -->
	</div>
{/snippet}
