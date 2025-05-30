<script lang="ts">
	import type { Component } from "svelte";
	import { v4 as uuidv4 } from "uuid";
	import type { IncidentEventEvidence, IncidentEventEvidenceAttributes } from "$lib/api";

	import { mdiPencil, mdiPlus, mdiSlack, mdiTrashCan, mdiWeb } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";
	import { Button, ListItem, SelectField, type MenuOption } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import Slack from "./data-sources/Slack.svelte";
	import Url from "./data-sources/Url.svelte";
	import { eventAttributes } from "./eventAttributesState.svelte";
	// import Github from "./data-sources/Github.svelte";

	type DataSourceComponent = Component<{ dataValue: string }, {}, "dataValue">;
	type DataSourceMenuOption = MenuOption<string> & { component: DataSourceComponent };
	const dataSourceOptions: DataSourceMenuOption[] = [
		{ value: "slack", label: "Slack", icon: mdiSlack, component: Slack },
		// { value: "github", label: "Github", icon: mdiGithub, component: Github },
		// { value: "metric", label: "Metric", icon: mdiMetric, component: Metric },
		// { value: "log", label: "Log", icon: mdiLog, component: Log },
		{ value: "url", label: "Web URL", icon: mdiWeb, component: Url },
	];

	let editing = $state<IncidentEventEvidence>();
	const editOption = $derived(
		editing ? dataSourceOptions.find((o) => o.value === editing?.attributes.source) : undefined
	);
	const setEditing = (ev: IncidentEventEvidence) => (editing = $state.snapshot(ev));
	const cancelEditing = () => (editing = undefined);
	const confirmEdit = () => {
		if (!editing) return;
		const idx = eventAttributes.evidence.findIndex((ev) => ev.id === editing?.id);
		if (idx === -1) return;
		eventAttributes.evidence[idx] = $state.snapshot(editing);
		editing = undefined;
	};

	let adding = $state<IncidentEventEvidenceAttributes>();
	const addOption = $derived(
		adding ? dataSourceOptions.find((o) => o.value === adding?.source) : undefined
	);

	const setAddingNew = () => (adding = { source: "", value: "" });
	const cancelAddingNew = () => (adding = undefined);
	const confirmAdd = () => {
		if (!adding) return;
		eventAttributes.evidence.push({ id: uuidv4(), attributes: $state.snapshot(adding) });
		adding = undefined;
	};

	const confirmDelete = (ev: IncidentEventEvidence) => {
		if (!confirm("Are you sure you want to delete this evidence?")) return;
		const idx = eventAttributes.evidence.findIndex((e) => e.id === ev.id);
		if (idx === -1) return;
		eventAttributes.evidence.splice(idx, 1);
	};
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	{#if adding}
		<SelectField bind:value={adding.source} options={dataSourceOptions} label="Data Source" />

		{#if addOption?.component}
			{@const SourceComponent = addOption.component}
			<SourceComponent bind:dataValue={adding.value} />
		{/if}

		<div class="w-full flex justify-end">
			<ConfirmButtons
				closeText="Cancel"
				onClose={cancelAddingNew}
				onConfirm={confirmAdd}
				saveEnabled={!!adding.value}
			/>
		</div>
	{:else if editing}
		{#if editOption?.component}
			{@const SourceComponent = editOption.component}
			<SourceComponent bind:dataValue={editing.attributes.value} />
		{/if}

		<div class="w-full flex justify-end">
			<ConfirmButtons
				closeText="Cancel"
				onClose={cancelEditing}
				onConfirm={confirmEdit}
				saveEnabled={!!editing.attributes.value}
			/>
		</div>
	{:else}
		{#each eventAttributes.evidence as ev, i}
			<ListItem
				title={ev.attributes.source}
				subheading={ev.attributes.value}
				classes={{ root: "border first:border-t rounded elevation-0" }}
				class="flex-1"
				noShadow
			>
				<div slot="actions">
					<Button icon={mdiPencil} iconOnly on:click={() => setEditing(ev)} />
					<Button icon={mdiTrashCan} iconOnly on:click={() => confirmDelete(ev)} />
				</div>
			</ListItem>
		{/each}

		<Button
			class="text-surface-content/50 p-2"
			color="primary"
			variant="fill-light"
			on:click={setAddingNew}
		>
			<span class="flex items-center gap-2 text-primary-content">
				Add Evidence
				<Icon data={mdiPlus} />
			</span>
		</Button>
	{/if}
</div>
