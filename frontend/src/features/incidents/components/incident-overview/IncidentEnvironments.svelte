<script lang="ts">
	import { createMutation } from "@tanstack/svelte-query";
	import {
		listEnvironmentsOptions,
		updateIncidentMutation,
		type Incident,
	} from "$lib/api";
	import ConfirmChangeButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { Button, Header, Icon } from "svelte-ux";
	import { mdiPencil } from "@mdi/js";
	import { incidentCtx } from "$features/incidents/lib/context";

	type Props = {};
	const {}: Props = $props();

	const incident = incidentCtx.get();

	let incidentEnvironments: string[] = [];
	let selectedEnvironments: string[] = $state([]);
	let changed = $state(false);
	let editing = $state(false);

	const resetState = (inc: Incident) => {
		const ids = inc.attributes.environments.map((e) => e.id);
		incidentEnvironments = [...ids];
		selectedEnvironments = [...ids];
		changed = false;
		editing = false;
	};
	$effect(() => resetState(incident));

	const update = createMutation(() => ({
		...updateIncidentMutation(),
		onSuccess: () => {
			resetState(incident);
		},
	}));
	const doEnvironmentUpdate = () => {
		update.mutate({
			path: { id: incident.id },
			body: { attributes: { environments: selectedEnvironments } },
		});
	};

	const arraysEqual = (a: string[], b: string[]) => {
		if (a === b) return true;
		if (a.length !== b.length) return false;
		const sb = b.toSorted();
		return a.toSorted().every((val, idx) => val === sb[idx]);
	};

	const onChanged = () =>
		(changed = !arraysEqual(selectedEnvironments, incidentEnvironments));
</script>

{#if !editing}
	<Header
		title="Impacted Environments"
		classes={{ root: "min-h-8", title: "text-md text-neutral-100" }}
	>
		<div slot="actions" class:hidden={editing}>
			<Button
				size="sm"
				classes={{ root: "h-8" }}
				on:click={() => {
					editing = true;
				}}
			>
				<span class="hidden group-hover:inline">Edit</span>
				<Icon data={mdiPencil} />
			</Button>
		</div>
	</Header>
	<div class="flex flex-col">
		{#each incident.attributes.environments as env}
			<span>
				{env.attributes.name}
			</span>
		{/each}
	</div>
{:else}
	<div class="flex flex-col gap-2 h-fit">
		<!--LoadingSelect
			id="incident-environments"
			label="Environments"
			options={listEnvironmentsOptions}
			bind:value={selectedEnvironments}
			multi
			disabled={update.isPending}
			on:change={() => {
				onChanged();
			}}
		/-->

		<ConfirmChangeButtons
			disabled={update.isPending}
			saveEnabled={changed}
			onClose={() => resetState(incident)}
			onConfirm={doEnvironmentUpdate}
		/>
	</div>
{/if}
