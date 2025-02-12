<script lang="ts">
	import { Collapse, Header, ListItem } from "svelte-ux";
	import { mdiGraphOutline, mdiLayers, mdiLink, mdiStateMachine } from "@mdi/js";
	import type { TimelineEvent } from "../types";

	import EventDetailsPanel from "./panels/EventDetails.svelte";
	import DecisionContextPanel from "./panels/DecisionContext.svelte";
	import ContributingFactorsPanel from "./panels/ContributingFactors.svelte";
	import EvidencePanel from "./panels/Evidence.svelte";
	import SystemComponentsPanel from "./panels/SystemComponents.svelte";
	import type { Component } from "svelte";
	import { eventAttributes } from "./panels/eventAttributes.svelte";
</script>

<div class="flex flex-row min-h-0 max-h-full h-full gap-2 p-2">
	<div class="flex flex-col gap-2">
		<Header title="Details" />

		<EventDetailsPanel />
	</div>

	<div class="flex flex-col gap-2 overflow-y-auto flex-1">
		<Header title="Context" />

		{#snippet panel(title: string, subheading: string, icon: string, PanelComponent: Component)}
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

		{#if eventAttributes.eventType === "decision"}
			{@render panel(
				"Decision Context",
				"Document the options, constraints, and reasoning behind this choice",
				mdiGraphOutline,
				DecisionContextPanel
			)}
		{/if}

		{@render panel(
			"Contributing Factors",
			"Identify pressures and conditions that shaped this event",
			mdiLayers,
			ContributingFactorsPanel
		)}

		{@render panel(
			"Evidence & Links",
			"Add links to logs, metrics, discussions, and other supporting information",
			mdiLink,
			EvidencePanel
		)}

		{@render panel(
			"System Context",
			"Document the relevant system components and their conditions at this time",
			mdiStateMachine,
			SystemComponentsPanel
		)}
	</div>
</div>
