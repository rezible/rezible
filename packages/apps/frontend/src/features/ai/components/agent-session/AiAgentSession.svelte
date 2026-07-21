<script lang="ts">
	import * as Collapsible from "$components/ui/collapsible";
	import { Badge } from "$components/ui/badge";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Spinner } from "$components/ui/spinner";
	import RiArrowRightSLine from "remixicon-svelte/icons/arrow-right-s-line";

	import AgentTurn from "./AgentTurn.svelte";
	import { initAgentSessionComponentController } from "./controller.svelte";

	type Props = {
		sessionId: string;
	};

	const { sessionId }: Props = $props();
	const ctrl = initAgentSessionComponentController(() => sessionId);
	const attrs = $derived(ctrl.session?.attributes);
	const status = $derived(attrs?.latestTurn?.attributes.status);
</script>

<section class="w-full min-w-0 space-y-3">
	{#if ctrl.error}
		<InlineAlert error={ctrl.error} dismissable={false} />
	{/if}

	<Collapsible.Root bind:open={() => ctrl.isExpanded, (open) => ctrl.setExpanded(open)} class="group">
		<div class="rounded-md border border-border bg-card text-card-foreground">
			<Collapsible.Trigger
				class="flex w-full min-w-0 items-center gap-3 p-4 text-left outline-none hover:bg-muted/40 focus-visible:ring-1 focus-visible:ring-ring"
			>
				<RiArrowRightSLine
					class="size-5 shrink-0 text-muted-foreground transition-transform group-data-[state=open]:rotate-90"
				/>
				<div class="min-w-0 flex-1">
					<h2 class="truncate text-sm font-semibold">{attrs?.agentName ?? "AI agent"}</h2>
					<p class="text-xs text-muted-foreground">{ctrl.turns.length} turns</p>
				</div>
				{#if status}
					<Badge variant={status === "failed" ? "destructive" : "outline"} class="capitalize"
						>{status}</Badge
					>
				{/if}
				{#if ctrl.isLoading}
					<Spinner aria-label="Agent session loading" />
				{/if}
			</Collapsible.Trigger>

			<Collapsible.Content>
				<div class="space-y-3 border-t border-border p-4">
					{#each ctrl.turns as turn (turn.id)}
						<AgentTurn {turn} />
					{:else}
						<p class="text-sm text-muted-foreground">No turns are available for this session.</p>
					{/each}
				</div>
			</Collapsible.Content>
		</div>
	</Collapsible.Root>
</section>
