<script lang="ts">
	import { mdiPencil } from "@mdi/js";
	import { Header, Button, Icon } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { useIncidentViewState } from "../viewState.svelte";

	const viewState = useIncidentViewState();

	const attrs = $derived(viewState.incident?.attributes);
	const teamAssignments = $derived(attrs?.teams ?? []);
	const roleAssignments = $derived(attrs?.roles ?? []);
	const linkedIncidents = $derived(attrs?.linkedIncidents ?? []);
</script>

<div class="flex flex-col gap-2 overflow-y-auto">
	<div class="grid grid-cols-2 gap-2">
		<div class="flex flex-col gap-2">
			<div class="border rounded-lg p-2 group">
				<Header title="Responders" classes={{ root: "min-h-8", title: "text-md text-neutral-100" }}>
					<div slot="actions">
						<Button size="sm" classes={{ root: "h-8 text-neutral-200" }} on:click={() => {}}>
							<Icon data={mdiPencil} />
						</Button>
					</div>
				</Header>
				
				{#each roleAssignments as assignment}
					<div class="">
						<span class="items-center flex flex-row gap-2">
							<Avatar kind="user" id={assignment.user.attributes.name} />
							<div class="flex flex-col">
								<span class="text-lg">{assignment.user.attributes.name}</span>
								<span class="text-gray-700">{assignment.role.attributes.name}</span>
							</div>
						</span>
					</div>
				{/each}
			</div>

			<div class="border rounded-lg p-2 group">
				<Header title="Teams" classes={{ root: "min-h-8", title: "text-md text-neutral-100" }}>
					<div slot="actions">
						<Button size="sm" classes={{ root: "h-8 text-neutral-200" }} on:click={() => {}}>
							<Icon data={mdiPencil} />
						</Button>
					</div>
				</Header>
				
				<div class="flex flex-col gap-2">
					{#each teamAssignments as assignment}
						{@const team = assignment.team}
						<div class="">
							<span class="items-center flex flex-row gap-2">
								<Avatar kind="team" id={team.id} />
								<div class="flex flex-col">
									<span class="text-lg">{team.attributes.name}</span>
								</div>
							</span>
						</div>
					{/each}
				</div>
			</div>

			<div class="border rounded-lg p-2 group">
				<Header title="Linked Incidents" classes={{ root: "min-h-8", title: "text-md text-neutral-100" }}>
					<div slot="actions">
						<Button size="sm" classes={{ root: "h-8 text-neutral-200" }} on:click={() => {}}>
							<Icon data={mdiPencil} />
						</Button>
					</div>
				</Header>
				
				<div class="flex flex-col gap-2">
					{#each linkedIncidents as linked}
						<a href="#/incidents/{linked.incidentId}">
							<div class="border p-2 hover:bg-accent cursor-pointer">
								<div class="text-lg">{linked.incidentTitle}</div>
								<div class="text-md">{linked.incidentSummary}</div>
							</div>
						</a>
					{:else}
						<span>no linked incidents</span>
					{/each}
				</div>
			</div>
		</div>

		<div class="flex flex-col gap-2">
			<div class="border rounded-lg p-2 group">
				<Header title="Incident Severity" classes={{ root: "h-8", title: "text-md text-neutral-100" }}>
					<div slot="actions">
						<Button size="sm" classes={{ root: "h-8 text-neutral-200" }} on:click={() => {}}>
							<Icon data={mdiPencil} />
						</Button>
					</div>
				</Header>
				<span>{attrs?.severity.attributes.name ?? "severity"}</span>
			</div>

			<div class="border rounded-lg p-2 group">
				<Header title="Incident Visibility" classes={{ root: "min-h-8", title: "text-md text-neutral-100" }}>
					<div slot="actions">
						<Button size="sm" classes={{ root: "h-8 text-neutral-200" }} on:click={() => {}}>
							<Icon data={mdiPencil} />
						</Button>
					</div>
				</Header>
				{#if attrs?.private}
					<span class="text-neutral-content">Restricted</span>
				{:else}
					<span class="text-neutral-content">Public</span>
				{/if}
			</div>
		</div>
	</div>
</div>
