<script lang="ts">
	import type { Component } from "svelte";
	import { v4 as uuidv4 } from "uuid";
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import { Button } from "$components/ui/button";
	import ConfirmButtons from "$components/forms/confirm-buttons/ConfirmButtons.svelte";
	import Slack from "./data-sources/Slack.svelte";
	import Url from "./data-sources/Url.svelte";
	import { useEventDialogAttributes } from "./attributes.svelte";
	import type { TimelineEntryEvidence, TimelineEntryEvidenceAttributes } from "../../entry-model";

	const attributes = useEventDialogAttributes();

	type MenuOption<T> = { label: string; value: T; };

	type DataSourceComponent = Component<{ dataValue: string }, Record<string, never>, "dataValue">;
	type DataSourceMenuOption = MenuOption<string> & { component: DataSourceComponent };
	const dataSourceOptions: DataSourceMenuOption[] = [
		{ value: "slack", label: "Slack", component: Slack },
		// { value: "github", label: "Github", icon: RiGithubFill, component: Github },
		// { value: "metric", label: "Metric", icon: RiLineChartLine, component: Metric },
		// { value: "log", label: "Log", icon: RiFileList3Line, component: Log },
		{ value: "url", label: "Web URL", component: Url },
	];

	let editing = $state<TimelineEntryEvidence>();
	const editOption = $derived(
		editing ? dataSourceOptions.find((o) => o.value === editing?.attributes.source) : undefined
	);
	const cancelEditing = () => (editing = undefined);
	const confirmEdit = () => {
		if (!editing) return;
		const idx = attributes.evidence.findIndex((ev) => ev.id === editing?.id);
		if (idx === -1) return;
		attributes.evidence[idx] = $state.snapshot(editing);
		editing = undefined;
	};

	let adding = $state<TimelineEntryEvidenceAttributes>();
	const addOption = $derived(
		adding ? dataSourceOptions.find((o) => o.value === adding?.source) : undefined
	);

	const setAddingNew = () => (adding = { source: "", value: "" });
	const cancelAddingNew = () => (adding = undefined);
	const confirmAdd = () => {
		if (!adding) return;
		attributes.evidence.push({ id: uuidv4(), attributes: $state.snapshot(adding) });
		adding = undefined;
	};
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	{#if adding}
		<span>data source select field</span>
		<!-- <SelectField bind:value={adding.source} options={dataSourceOptions} label="Data Source" /> -->

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
		{#each attributes.evidence as ev (ev.id)}
			<span>evidence: {ev.attributes.source}</span>
			<!-- <ListItem
				title={ev.attributes.source}
				subheading={ev.attributes.value}
				classes={{ root: "border first:border-t rounded elevation-0" }}
				class="flex-1"
				noShadow
			>
				<div slot="actions">
					<Button icon={RiPencilLine} iconOnly onclick={() => setEditing(ev)} />
					<Button icon={RiDeleteBinLine} iconOnly onclick={() => confirmDelete(ev)} />
				</div>
			</ListItem> -->
		{/each}

		<Button color="primary" onclick={setAddingNew}>
			<span class="flex items-center gap-2 text-primary-content">
				Add Evidence
				<RiAddLine class="" aria-hidden="true" />
			</span>
		</Button>
	{/if}
</div>
