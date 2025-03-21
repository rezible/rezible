<script lang="ts">
	import type { Incident, User } from "$lib/api";

	type Props = {
		user: User;
		incidents: Incident[];
	};
	const { user, incidents }: Props = $props();

	// Calculate stats
	const totalIncidents = incidents.length;
	const resolvedIncidents = incidents.filter((inc) => inc.attributes.currentStatus === "resolved").length;
	const leadIncidents = incidents.filter((inc) =>
		inc.attributes.roles?.some((r) => r.user.id === user.id && r.role.attributes.name === "incident_lead")
	).length;

	// Get recent incidents (last 5)
	const recentIncidents = [...incidents]
		.sort((a, b) => new Date(b.attributes.openedAt).getTime() - new Date(a.attributes.openedAt).getTime())
		.slice(0, 5);

	// Calculate severity distribution
	const severityCount = incidents.reduce<Record<string, number>>((acc, inc) => {
		const sevId = inc.attributes.severity.id;
		acc[sevId] = (acc[sevId] || 0) + 1;
		return acc;
	}, {});
</script>

<div class="">
	<h2>Incident Participation</h2>

	<div class="">
		<div class="">
			<div class="">{totalIncidents}</div>
			<div class="">Total Incidents</div>
		</div>

		<div class="">
			<div class="">{leadIncidents}</div>
			<div class="">Led as IC</div>
		</div>

		<div class="">
			<div class="">{Math.round((resolvedIncidents / totalIncidents) * 100) || 0}%</div>
			<div class="">Resolution Rate</div>
		</div>
	</div>

	{#if Object.keys(severityCount).length > 0}
		<div class="">
			<h3>Severity Distribution</h3>
			<div class="">
				{#each Object.entries(severityCount) as [severity, count]}
					<div class="">
						<div class="">{severity}</div>
						<div class="">
							<div
								class=""
								style="width: {(count / totalIncidents) * 100}%"
							>
								<span class="">{count}</span>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	{#if recentIncidents.length > 0}
		<div class="">
			<h3>Recent Incidents</h3>
			<ul>
				{#each recentIncidents as incident}
					{@const attr = incident.attributes}
					<li>
						<a href="/incidents/{attr.slug}" class="">
							<div class="">
								<span class="">{attr.title}</span>
							</div>
							<div class="">
								{new Date(attr.openedAt).toLocaleDateString()}
								{#if attr.roles?.some((r) => r.user.id === user.id)}
									• {attr.roles?.find((r) => r.user.id === user.id)?.role.attributes.name.replace("_", " ")}
								{/if}
							</div>
						</a>
					</li>
				{/each}
			</ul>
		</div>
	{/if}
</div>
