<script lang="ts">
	import { resolve } from "$app/paths";
	import { Button } from "$components/ui/button";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { useIncidentContextSidebarController } from "./controller.svelte";

	const controller = useIncidentContextSidebarController();
</script>

<div class="flex min-h-0 flex-col gap-3 p-2">
	<div class="flex items-center justify-between gap-2">
		<div class="min-w-0">
			<h3 class="text-sm font-semibold text-foreground">Triage context</h3>
			{#if controller.latestRunStatusLabel}
				<p class="truncate text-xs text-muted-foreground">Agent: {controller.latestRunStatusLabel}</p>
			{/if}
		</div>
		<Button
			size="sm"
			variant="outline"
			onclick={controller.requestContextPack}
			disabled={controller.requestDisabled}
		>
			{#if controller.requestPending}
				<Spinner />
				Run
			{:else}
				Refresh
			{/if}
		</Button>
	</div>

	<LoadingQueryWrapper query={controller.casesQuery}>
		{#snippet view()}
			<div class="grid gap-3">
				<section class="grid gap-2 border border-border p-2">
					<div class="flex items-center justify-between gap-2">
						<h4 class="text-xs font-semibold uppercase text-muted-foreground">
							Impacted systems
						</h4>
						<span class="text-xs text-muted-foreground">{controller.impacts.length}</span>
					</div>
					{#if controller.impacts.length > 0}
						<div class="grid gap-2">
							{#each controller.impactItems as impact (impact.id)}
								<div class="grid gap-1 border-l border-border pl-2">
									<span class="truncate text-sm font-medium">{impact.displayName}</span>
									<span class="text-xs text-muted-foreground"
										>{impact.kind} · {impact.score}</span
									>
									{#if impact.reason}
										<p class="line-clamp-2 text-xs leading-5 text-muted-foreground">
											{impact.reason}
										</p>
									{/if}
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-muted-foreground">No impacted systems identified.</p>
					{/if}
				</section>

				<section class="grid gap-2 border border-border p-2">
					<div class="flex items-center justify-between gap-2">
						<h4 class="text-xs font-semibold uppercase text-muted-foreground">Active alerts</h4>
						<span class="text-xs text-muted-foreground">{controller.activeAlerts.length}</span>
					</div>
					{#if controller.activeAlerts.length > 0}
						<div class="grid gap-2">
							{#each controller.activeAlertItems as alert (alert.id)}
								<div class="grid gap-1">
									<a
										class="truncate text-sm font-medium hover:underline"
										href={resolve("/alerts/[id]/[[view=alertView]]", {
											id: alert.routeId,
										})}
									>
										{alert.title}
									</a>
									{#if alert.summary}
										<p class="line-clamp-2 text-xs leading-5 text-muted-foreground">
											{alert.summary}
										</p>
									{/if}
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-muted-foreground">No active alerts linked.</p>
					{/if}
				</section>

				<section class="grid gap-2 border border-border p-2">
					<div class="flex items-center justify-between gap-2">
						<h4 class="text-xs font-semibold uppercase text-muted-foreground">Prior incidents</h4>
						<span class="text-xs text-muted-foreground">{controller.relatedIncidents.length}</span
						>
					</div>
					{#if controller.relatedIncidents.length > 0}
						<div class="grid gap-2">
							{#each controller.relatedIncidentItems as incident (incident.id)}
								<div class="grid gap-1">
									<a
										class="truncate text-sm font-medium hover:underline"
										href={resolve("/incidents/[slug]/[[view=incidentView]]", {
											slug: incident.routeSlug,
										})}
									>
										{incident.title}
									</a>
									{#if incident.summary}
										<p class="line-clamp-2 text-xs leading-5 text-muted-foreground">
											{incident.summary}
										</p>
									{/if}
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-muted-foreground">No related incidents found.</p>
					{/if}
				</section>
			</div>
		{/snippet}
	</LoadingQueryWrapper>
</div>
