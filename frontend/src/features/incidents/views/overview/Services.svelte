<script lang="ts">
	import type { Incident, IncidentResourceImpactService, Service } from '$lib/api';
    import Avatar from '$components/avatar/Avatar.svelte';

	interface Props { incident: Incident };
	let { incident }: Props = $props();

	const serviceImpacts = $derived(incident.attributes.impacted_services);
</script>

<span class="text-neutral-100">Services</span>

<div class="flex flex-col gap-2">
	{#each serviceImpacts as impact}
		{@const service = impact.resource}
		<div class="">
			<span class="items-center flex flex-row gap-2">
				<Avatar kind="service" id={service.id} />
				<div class="flex flex-col">
					<div class="text-lg">{service.attributes.name}</div>
					<div class="text-md">{impact.summary}</div>
				</div>
				{#if !!impact.incident_id}
					<a href="/incidents/{impact.incident_id}" class="hover:bg-accent cursor-pointer">
						View Incident
					</a>
				{/if}
			</span>
		</div>
	{/each}
</div>
