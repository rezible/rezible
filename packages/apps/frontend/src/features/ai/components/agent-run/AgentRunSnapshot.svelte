<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import { Spinner } from "$components/ui/spinner";

	import AgentRunArtifact from "./AgentRunArtifact.svelte";
	import AgentRunMessage from "./AgentRunMessage.svelte";
	import type { AgentRunSnapshotView } from "./controller.svelte";

	type Props = {
		snapshot: AgentRunSnapshotView;
	};

	const { snapshot }: Props = $props();
</script>

<article class="space-y-4 rounded border border-border bg-card p-4 text-card-foreground">
	<header class="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
		<div class="min-w-0 space-y-1">
			<div class="flex flex-wrap items-center gap-2">
				<Badge variant={snapshot.statusVariant} class={snapshot.statusClass}
					>{snapshot.statusLabel}</Badge
				>
				{#if snapshot.isPending}
					<Spinner aria-label="Snapshot pending" />
				{/if}
			</div>
			<p class="break-all text-xs text-muted-foreground">{snapshot.id}</p>
		</div>
		<div class="text-left text-xs text-muted-foreground sm:text-right">{snapshot.createdAt}</div>
	</header>

	<dl class="grid gap-3 text-sm sm:grid-cols-2 lg:grid-cols-4">
		<div class="min-w-0">
			<dt class="font-medium text-muted-foreground">State</dt>
			<dd class="text-foreground">{snapshot.stateSummary}</dd>
		</div>
		<div class="min-w-0">
			<dt class="font-medium text-muted-foreground">Finish reason</dt>
			<dd class="break-words text-foreground">{snapshot.finishReason}</dd>
		</div>
		<div class="min-w-0">
			<dt class="font-medium text-muted-foreground">Parent</dt>
			<dd class="break-all text-foreground">{snapshot.parentId}</dd>
		</div>
		<div class="min-w-0">
			<dt class="font-medium text-muted-foreground">Heartbeat</dt>
			<dd class="text-foreground">{snapshot.heartbeatAt}</dd>
		</div>
	</dl>

	{#if snapshot.error}
		<div class="rounded border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
			{snapshot.error}
		</div>
	{/if}

	{#if snapshot.stateCustom}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Custom state</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-background p-3 text-xs leading-relaxed text-foreground">{snapshot.stateCustom}</pre>
		</div>
	{/if}

	<section class="space-y-3">
		<h3 class="text-sm font-medium text-foreground">Messages</h3>
		<div class="space-y-3">
			{#each snapshot.messages as message (message)}
				<AgentRunMessage {message} />
			{:else}
				<p class="text-sm text-muted-foreground">No messages.</p>
			{/each}
		</div>
	</section>

	<section class="space-y-3">
		<h3 class="text-sm font-medium text-foreground">Artifacts</h3>
		<div class="space-y-3">
			{#each snapshot.artifacts as artifact (artifact.name)}
				<AgentRunArtifact {artifact} />
			{:else}
				<p class="text-sm text-muted-foreground">No artifacts.</p>
			{/each}
		</div>
	</section>
</article>
