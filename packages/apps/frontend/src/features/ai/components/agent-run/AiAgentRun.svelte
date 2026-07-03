<script lang="ts">
	import * as Collapsible from "$components/ui/collapsible";
	import { Badge } from "$components/ui/badge";
	import { Spinner } from "$components/ui/spinner";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import RiArrowRightSLine from "remixicon-svelte/icons/arrow-right-s-line";

	import AgentRunSnapshot from "./AgentRunSnapshot.svelte";
	import { initAiAgentRunComponentController } from "./controller.svelte";

	type Props = {
		runId: string;
	};

	const { runId }: Props = $props();
	const ctrl = initAiAgentRunComponentController(() => runId);
</script>

<section class="w-full min-w-0 space-y-3">
	{#if ctrl.error}
		<InlineAlert error={ctrl.error} dismissable={false} />
	{/if}

	<Collapsible.Root bind:open={() => ctrl.isExpanded, (open) => ctrl.setExpanded(open)} class="group">
		<div class="rounded border border-border bg-card text-card-foreground">
			<Collapsible.Trigger
				class="flex w-full min-w-0 items-start gap-3 p-4 text-left outline-none transition-colors hover:bg-muted/40 focus-visible:ring-1 focus-visible:ring-ring"
			>
				<RiArrowRightSLine
					class="mt-0.5 size-5 shrink-0 text-muted-foreground transition-transform group-data-[state=open]:rotate-90"
				/>

				<div class="min-w-0 flex-1 space-y-2">
					<div class="flex min-w-0 flex-wrap items-center gap-2">
						<h2 class="truncate text-sm font-semibold text-foreground">{ctrl.agentName}</h2>
						{#if ctrl.latestSnapshot}
							<Badge
								variant={ctrl.latestSnapshot.statusVariant}
								class={ctrl.latestSnapshot.statusClass}
							>
								{ctrl.latestSnapshot.statusLabel}
							</Badge>
						{:else}
							<Badge variant="outline">No snapshots</Badge>
						{/if}
						{#if ctrl.isLoading || ctrl.isLatestPending}
							<Spinner aria-label="Agent run pending" />
						{/if}
					</div>

					<div class="grid gap-1 text-sm text-muted-foreground sm:grid-cols-2">
						<div class="min-w-0">
							<span class="font-medium text-foreground">State:</span>
							<span>{ctrl.latestStateSummary}</span>
						</div>
						{#if ctrl.latestSnapshot}
							<div class="min-w-0 sm:text-right">
								<span class="font-medium text-foreground">Updated:</span>
								<span>{ctrl.latestSnapshot.createdAt}</span>
							</div>
						{/if}
					</div>
				</div>
			</Collapsible.Trigger>

			<Collapsible.Content>
				<div class="border-t border-border p-4">
					<div class="space-y-4">
						{#each ctrl.snapshots as snapshot (snapshot.id)}
							<AgentRunSnapshot {snapshot} />
						{:else}
							<p class="text-sm text-muted-foreground">
								No snapshots are available for this agent run.
							</p>
						{/each}
					</div>
				</div>
			</Collapsible.Content>
		</div>
	</Collapsible.Root>
</section>
