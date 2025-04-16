<script lang="ts">
	import type { Component } from "svelte";
	import { Collapse, Header, ListItem } from "svelte-ux";
	import { mdiGraphOutline, mdiLayers, mdiLink, mdiStateMachine } from "@mdi/js";

	import { eventAttributes } from "./attribute-panels/eventAttributesState.svelte";
	import EventDetailsPanel from "./attribute-panels/EventDetails.svelte";
	import DecisionContextPanel from "./attribute-panels/DecisionContext.svelte";
	import ContributingFactorsPanel from "./attribute-panels/ContributingFactors.svelte";
	import EvidencePanel from "./attribute-panels/Evidence.svelte";
	import SystemContextPanel from "./attribute-panels/SystemContext.svelte";
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
			{#if eventAttributes.kind === "decision"}
				{@render componentTraitPanel(
					"Decision Context",
					"Document the options, constraints, and reasoning behind this choice",
					mdiGraphOutline,
					DecisionContextPanel
				)}
			{/if}

			{@render componentTraitPanel(
				"Contributing Factors",
				"Identify pressures and conditions that shaped this event",
				mdiLayers,
				ContributingFactorsPanel
			)}

			{@render componentTraitPanel(
				"Evidence & Links",
				"Add links to logs, metrics, discussions, and other supporting information",
				mdiLink,
				EvidencePanel
			)}

			{@render componentTraitPanel(
				"System Context",
				"Document the relevant system components and their conditions at this time",
				mdiStateMachine,
				SystemContextPanel
			)}
		</div>
	</div>
</div>

{#snippet componentTraitPanel(title: string, subheading: string, icon: string, PanelComponent: Component)}
<div class="p-2 border rounded">
	<Collapse open classes={{ root: "overflow-x-hidden", content: "p-2" }}>
		<ListItem
			slot="trigger"
			{title}
			{subheading}
			{icon}
			classes={{ root: "pl-0" }}
			avatar={{
				class: "bg-surface-content/50 text-surface-100/90",
			}}
			class="flex-1"
			noShadow
		/>
		<PanelComponent></PanelComponent>
	</Collapse>
</div>
{/snippet}