<script lang="ts">
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Badge, type BadgeVariant } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import * as Collapsible from "$components/ui/collapsible";
	import { Spinner } from "$components/ui/spinner";
	import type { ProposalStatusView, ProposalView } from "$features/system/lib/catalogue/proposals";
	import { useCatalogueController } from "../controller.svelte";
	import AliasList from "../alias-list/AliasList.svelte";

	type Props = { proposal: ProposalView };
	const { proposal }: Props = $props();

	const controller = useCatalogueController();

	const statusBadgeVariants: Record<ProposalStatusView["tone"], BadgeVariant> = {
		pending: "info",
		rejected: "neutral",
		confirmed: "success",
	};

	const rejecting = $derived(controller.rejectingId === proposal.id);
	const rejectError = $derived(controller.rejectErrorFor(proposal.id));
</script>

<article aria-label={`${proposal.status.label}: ${proposal.label}`}>
	<Card.Root class="gap-3 py-4">
		<Card.Header class="px-4">
			<div class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
				<Badge variant={statusBadgeVariants[proposal.status.tone]}>{proposal.status.label}</Badge>
				<span>{proposal.origin}</span>
				<span title={proposal.createdTitle}>· proposed {proposal.created}</span>
			</div>
		</Card.Header>

		<Card.Content class="space-y-3 px-4">
			{#if proposal.sharedEntity}
				<div class="min-w-0 space-y-1 rounded-md border p-3">
					<p class="text-sm">
						{#each proposal.values as value, index (value.key)}
							{#if index > 0}
								{" and "}
							{/if}
							<code class="font-medium wrap-anywhere">{value.value}</code>
						{/each}
						{` are one ${proposal.sharedEntity.kind}.`}
					</p>
					<AliasList aliases={proposal.sharedEntity.aliases} />
				</div>
			{:else}
				<div class="grid gap-3 sm:grid-cols-2">
					{#each proposal.values as value (value.key)}
						<div class="min-w-0 space-y-1 rounded-md border p-3">
							<div class="flex flex-wrap items-baseline gap-x-2">
								<code class="text-sm font-medium wrap-anywhere">{value.value}</code>
								{#if value.displayName}
									<span class="text-xs text-muted-foreground">
										shown as {value.displayName}
									</span>
								{/if}
							</div>
							{#if value.unheld}
								<p class="text-xs text-muted-foreground">
									No longer names anything Rezible knows.
								</p>
							{:else}
								<AliasList aliases={value.aliases} />
							{/if}
						</div>
					{/each}
				</div>
			{/if}

			<div class="text-sm">
				<span class="text-xs font-medium text-muted-foreground">Rezible's suggestion</span>
				<p>
					{#each proposal.rationale as part (part.key)}
						{#if part.code}
							<code>{part.text}</code>
						{:else}
							{part.text}
						{/if}
					{/each}
				</p>
			</div>

			<Collapsible.Root>
				<Collapsible.Trigger
					class="group flex items-center gap-1 text-xs font-medium text-muted-foreground hover:text-foreground"
				>
					<RiArrowDownSLine
						class="size-4 transition-transform group-data-[state=open]:rotate-180"
					/>
					Cited evidence
				</Collapsible.Trigger>
				<Collapsible.Content class="pt-1 pl-5 text-xs text-muted-foreground">
					<p>
						<span class="font-medium text-foreground">{proposal.evidence.integration}</span>
						<span>named it</span>
						<code class="text-foreground">{proposal.evidence.name}</code>
						<span title={proposal.evidence.recordedTitle}>{proposal.evidence.recorded}</span>
					</p>
					<p>
						{proposal.evidence.assertion} ·
						<code>{proposal.evidence.resourceRef}</code>
					</p>
				</Collapsible.Content>
			</Collapsible.Root>

			{#if proposal.decisions.length > 0}
				<ul class="space-y-0.5 text-xs text-muted-foreground" aria-label="Decisions">
					{#each proposal.decisions as decision (decision.key)}
						<li>
							{decision.text}
							<span title={decision.timeTitle}>· {decision.time}</span>
						</li>
					{/each}
				</ul>
			{/if}

			{#if rejectError}
				<InlineAlert error={rejectError} dismissable={false} />
			{/if}
		</Card.Content>

		{#if proposal.status.actions.length > 0}
			<Card.Footer class="gap-2 px-4">
				{#if proposal.status.actions.includes("confirm")}
					<Button size="sm" onclick={() => controller.requestConfirm(proposal.id)}>Confirm</Button>
				{/if}
				{#if proposal.status.actions.includes("reject")}
					<Button
						size="sm"
						variant="outline"
						disabled={rejecting}
						onclick={() => controller.reject(proposal.id)}
					>
						{#if rejecting}
							<Spinner />
						{/if}
						Reject
					</Button>
				{/if}
			</Card.Footer>
		{/if}
	</Card.Root>
</article>
