<script lang="ts">
	import { mdiPencil } from "@mdi/js";
	import { Button } from "$components/ui/button";
	import Icon from "$components/icon/Icon.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import Header from "$components/header/Header.svelte";
	import { useIncidentViewController } from "$features/incidents/views/incident";

	const view = useIncidentViewController();

	const attrs = $derived(view.incident?.attributes);
	const teamAssignments = $derived(attrs?.teams ?? []);
	const roleAssignments = $derived(attrs?.roles ?? []);
	const linkedIncidents = $derived(attrs?.linkedIncidents ?? []);
</script>

<div class="flex flex-col gap-2 overflow-y-auto">
	<div class="grid grid-cols-2 gap-2">
		<div class="flex flex-col gap-2">
			<div class="border rounded-lg p-2 group">
				<Header title="Responders" classes={{ root: "min-h-8", title: "text-md text-neutral-100" }}>
					{#snippet actions()}
						<Button size="sm" onclick={() => {}}>
							<Icon data={mdiPencil} />
						</Button>
					{/snippet}
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
					{#snippet actions()}
						<Button size="sm" onclick={() => {}}>
							<Icon data={mdiPencil} />
						</Button>
					{/snippet}
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
					{#snippet actions()}
						<Button size="sm" onclick={() => {}}>
							<Icon data={mdiPencil} />
						</Button>
					{/snippet}
				</Header>
				
				<div class="flex flex-col gap-2">
					{#each linkedIncidents as linked}
						<a href="#/incidents/{linked.incidentId}">
							<div class="border p-2 hover:bg-accent cursor-pointer">
								<div class="text-lg">{linked.incidentTitle}</div>
								<div class="text-md">{linked.incidentSummary}</div>
							</div>
						</a>
					{/each}
				</div>
			</div>
		</div>

		<div class="flex flex-col gap-2">
			<div class="border rounded-lg p-2 group">
				<Header title="Incident Severity" classes={{ root: "h-8", title: "text-md text-neutral-100" }}>
					{#snippet actions()}
						<Button size="sm" onclick={() => {}}>
							<Icon data={mdiPencil} />
						</Button>
					{/snippet}
				</Header>
				<span>{attrs?.severity.attributes.name ?? "severity"}</span>
			</div>

			<div class="border rounded-lg p-2 group">
				<Header title="Incident Visibility" classes={{ root: "min-h-8", title: "text-md text-neutral-100" }}>
					{#snippet actions()}
						<Button size="sm" onclick={() => {}}>
							<Icon data={mdiPencil} />
						</Button>
					{/snippet}
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
