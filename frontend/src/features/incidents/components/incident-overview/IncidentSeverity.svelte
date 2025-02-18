<script lang="ts">
	import { createMutation, useQueryClient } from "@tanstack/svelte-query";
	import { Button, Icon, Header } from "svelte-ux";
	import { mdiPencil, mdiPencilOutline } from "@mdi/js";
	import { updateIncidentMutation, type Incident } from "$lib/api";
	import ConfirmChangeButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { incidentCtx } from "$features/incidents/lib/context";

	type Props = {};
	const {}: Props = $props();

	const incident = incidentCtx.get();

	let editing = $state(false);
	let newSeverity = $state(incident.attributes.severity.id);
	const changed = $derived(newSeverity !== incident.attributes.severity.id);

	const resetState = (inc: Incident) => {
		newSeverity = inc.attributes.severity.id;
		editing = false;
	};
	$effect(() => {
		resetState(incident);
	});

	const update = createMutation(() => ({
		...updateIncidentMutation(),
		onSuccess: () => {
			resetState(incident);
		},
	}));
	const doSeverityUpdate = () => {
		update.mutate({
			path: { id: incident.id },
			body: { attributes: { severityId: newSeverity } },
		});
	};
</script>

{#if !editing}
	<Header title="Incident Severity" classes={{ root: "h-8", title: "text-md text-neutral-100" }}>
		<div slot="actions" class:hidden={editing}>
			<Button
				size="sm"
				classes={{ root: "h-8" }}
				on:click={() => {
					editing = true;
				}}
			>
				<Icon data={mdiPencil} />
			</Button>
		</div>
	</Header>
	<span>{incident.attributes.severity.attributes.name}</span>
{:else}
	<div class="flex flex-col gap-2 w-64">
		<!--LoadingSelect
			label="Incident Severity"
			id="select-incident-severity"
			bind:value={newSeverity}
			options={incidentSeverities.list}
		/-->

		<ConfirmChangeButtons
			disabled={update.isPending}
			saveEnabled={changed}
			onClose={() => resetState(incident)}
			onConfirm={doSeverityUpdate}
		/>
	</div>
{/if}
