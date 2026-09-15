<script lang="ts">
	import "./styles.css";
	import { useIncidentView } from "$features/incidents/views/incident";

	import { useIncidentCollaboration } from "$features/incidents/views/incident/collaboration.svelte";

	import FieldEditorWrapper from "./field-editor/FieldEditorWrapper.svelte";
	import LoadingIndicator from "$src/components/layout/loading-indicator/LoadingIndicator.svelte";
	import { Button } from "$components/ui/button";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { initIncidentReportController } from "./controller.svelte";

	const incident = useIncidentView();
	const collab = useIncidentCollaboration();
	const report = initIncidentReportController();

	const docSessionQuery = $derived(incident.retrospectiveDocumentSessionQuery);
	const focusSectionFn = $state<Record<string, VoidFunction>>({});
</script>

{#snippet errorWithRetry(message: string, retryFn: () => void)}
	<div
		class="mb-4 flex items-center justify-between gap-3 rounded-md border border-destructive p-3 text-sm text-destructive"
		role="alert"
	>
		<span>{message}</span>
		<Button
			variant="outline"
			size="sm"
			onclick={retryFn}
		>Retry</Button>
	</div>
{/snippet}

{#if !report.retrospectiveId}
	<p class="p-6 text-sm text-muted-foreground">No report is associated with this incident.</p>
{:else}
	<LoadingQueryWrapper query={incident.retrospectiveQuery}>
		{#snippet view()}
			<div class="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-6">
				{#if docSessionQuery.isError && !docSessionQuery.data}
					{@render errorWithRetry("Unable to load report access.", docSessionQuery.refetch)}
				{:else if docSessionQuery.isPending && !docSessionQuery.data}
					<div class="flex items-center gap-3 p-4">
						<LoadingIndicator />
						<span>Loading report access…</span>
					</div>
				{:else if !report.canView}
					<p class="rounded-md border border-destructive p-4 text-sm text-destructive" role="alert">
						You do not have access to this report.
					</p>
				{:else}
					{#if docSessionQuery.isError}
						{@render errorWithRetry("Report access refresh failed.", docSessionQuery.refetch)}
					{/if}
					<div class="grid min-h-0 gap-6 lg:grid-cols-[minmax(0,1fr)_260px]">
						<div class="min-w-0 rounded-lg border border-border bg-card p-6">
							<div class="w-full overflow-y-auto flex flex-col gap-4">
								{#if !incident.retrospective}
									<p class="rounded-md border border-border p-6 text-sm text-muted-foreground">
										No report is associated with this incident.
									</p>
								{:else if !!collab.provider}
									{#if collab.error}
										{@render errorWithRetry(collab.error.message, collab.retry)}
									{/if}

									{#if !collab.initialSynced}
										<p class="mb-3 text-xs text-muted-foreground" role="status">
											Connecting...
										</p>
									{:else if collab.status === "disconnected"}
										{@render errorWithRetry("Report connection lost.", collab.retry)}
									{/if}

									{#each report.sections as section (section.field)}
										<div id={section.field}>
											<FieldEditorWrapper
												{section}
												bind:focusEditor={focusSectionFn[section.field]}
											/>
										</div>
									{/each}
								{:else if collab.error}
									{@render errorWithRetry(collab.error.message, collab.retry)}
								{:else}
									<div class="flex items-center gap-4">
										<LoadingIndicator />
										<span>Loading document...</span>
									</div>
								{/if}
							</div>
						</div>
						
						<aside class="rounded-lg border border-border bg-card p-5">
							<h2 class="font-semibold">Review</h2>
							<p class="mt-3 text-sm text-muted-foreground">
								{incident.retrospective?.attributes.state}
							</p>
							<h2 class="mt-6 font-semibold">In this report</h2>
							<nav class="mt-4 flex flex-col gap-3">
								{#each report.sections as section (section.field)}
									<a href={`#${section.field}`} class="text-sm text-muted-foreground hover:text-foreground">
										{section.title}
									</a>
								{/each}
							</nav>
						</aside>
					</div>
				{/if}
			</div>
		{/snippet}
	</LoadingQueryWrapper>
{/if}
