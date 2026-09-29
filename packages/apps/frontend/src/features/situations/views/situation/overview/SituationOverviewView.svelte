<script lang="ts">
	import * as Collapsible from "$components/ui/collapsible";
	import * as Empty from "$components/ui/empty";
	import * as Card from "$components/ui/card";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";
	import RiFileListLine from "remixicon-svelte/icons/file-list-line";
	import { useSituationController } from "../controller.svelte";
	import { initSituationInvestigationController } from "../investigation/controller.svelte";
	import { initSituationOverviewController } from "./controller.svelte";
	import SituationSourceSheet from "$features/situations/components/situation-source-sheet/SituationSourceSheet.svelte";
	import { investigationExecution, reportExcerpt, timestamp } from "../model";

	const sitCtrl = useSituationController();

	// TODO: not this
	const investigationController = initSituationInvestigationController(() => sitCtrl.investigationId);
	sitCtrl.setOverviewRefreshInterval(() => investigationController.refreshInterval);

	const controller = initSituationOverviewController();
	const attrs = $derived(controller.attrs);
	const investigationAttrs = $derived(investigationController.investigation?.attributes);
	const execution = $derived(investigationExecution(investigationAttrs));
	const openedAt = $derived(timestamp(attrs?.openedAt));
	const closedAt = $derived(timestamp(attrs?.closedAt));
	const closeReason = $derived(
		attrs?.closeReason === "stabilized"
			? "Stabilized"
			: attrs?.closeReason === "dismissed"
				? "Dismissed"
				: undefined
	);
	const report = $derived(investigationController.report?.attributes);
	const reportPublishedAt = $derived(timestamp(report?.createdAt));
	const excerpt = $derived(report ? reportExcerpt(report.text) : undefined);
</script>

{#if attrs}
	<div class="min-h-0 min-w-0 flex-1 overflow-y-auto p-4" tabindex="-1">
		<div class="mx-auto flex max-w-6xl flex-col gap-6">
			<section class="flex max-w-[75ch] flex-col gap-3" aria-labelledby="situation-title">
				<div class="flex flex-wrap items-center gap-3">
					<h1 id="situation-title" class="text-[28px] leading-9 font-semibold wrap-anywhere">
						{attrs.title}
					</h1>
					<Badge variant={attrs.closedAt ? "secondary" : "outline"}>
						{attrs.closedAt ? "Closed" : "Open"}
					</Badge>
				</div>
				{#if attrs.summary}
					<p
						class="whitespace-pre-wrap text-sm leading-relaxed text-muted-foreground wrap-anywhere"
					>
						{attrs.summary}
					</p>
				{/if}
				<dl class="flex flex-wrap gap-x-6 gap-y-2 text-xs text-muted-foreground">
					<div class="flex gap-1">
						<dt>Opened</dt>
						<dd><time datetime={openedAt.iso}>{openedAt.label}</time></dd>
					</div>
					{#if attrs.closedAt}
						<div class="flex gap-1">
							<dt>Closed</dt>
							<dd><time datetime={closedAt.iso}>{closedAt.label}</time></dd>
						</div>
					{/if}
					{#if closeReason}
						<div class="flex gap-1">
							<dt>Close reason</dt>
							<dd>{closeReason}</dd>
						</div>
					{/if}
				</dl>
			</section>

			<Card.Root>
				<Card.Header class="flex flex-row flex-wrap items-start justify-between gap-3">
					<div>
						<Card.Title>Current account</Card.Title>
						<Card.Description>Latest investigation report excerpt</Card.Description>
					</div>
					{#if controller.investigationId && !controller.investigationUnavailable && investigationAttrs}
						<Badge variant="secondary">{execution.label}</Badge>
					{:else if !controller.investigationId || controller.investigationUnavailable}
						<Badge variant="outline">Unavailable</Badge>
					{/if}
				</Card.Header>
				<Card.Content class="flex flex-col gap-4">
					{#if controller.investigationId && !controller.investigationUnavailable && investigationAttrs?.hasPendingWork}
						<p class="text-sm text-muted-foreground">Follow-up work is waiting.</p>
					{/if}
					{#if investigationController.investigationQuery.isError && !controller.investigationUnavailable}
						<div
							role={investigationAttrs ? "status" : "alert"}
							class="flex items-center justify-between gap-3 text-xs text-muted-foreground"
						>
							<span>
								{investigationAttrs
									? "Refresh failed. Showing previously loaded investigation details."
									: "Investigation details could not be loaded."}
							</span>
							<Button
								variant="ghost"
								size="sm"
								onclick={() => investigationController.investigationQuery.refetch()}
							>
								Retry
							</Button>
						</div>
					{/if}
					{#if !controller.investigationId || controller.investigationUnavailable}
						<p class="text-sm text-muted-foreground">Investigation is unavailable.</p>
					{:else if report && excerpt !== undefined}
						{#if investigationController.reportQuery.isError}
							<div
								role="status"
								class="flex items-center justify-between gap-3 text-xs text-muted-foreground"
							>
								<span>Refresh failed. Showing the previously loaded report.</span>
								<Button
									variant="ghost"
									size="sm"
									onclick={() => investigationController.reportQuery.refetch()}
								>
									Retry
								</Button>
							</div>
						{/if}
						<div class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
							<time datetime={reportPublishedAt.iso}>Published {reportPublishedAt.label}</time>
							{#if report.turnStatus === "running"}
								<Badge variant="outline">In progress</Badge>
							{/if}
						</div>
						<p class="max-w-[75ch] whitespace-pre-wrap text-[15px] leading-6 wrap-anywhere">
							{excerpt}
						</p>
					{:else if investigationController.reportAccessLost}
						<p class="text-sm text-muted-foreground">The latest report is unavailable.</p>
					{:else if investigationController.investigationQuery.isPending && !investigationAttrs}
						<p role="status" class="flex items-center gap-2 text-sm text-muted-foreground">
							<Spinner /> Loading investigation
						</p>
					{:else if investigationController.reportQuery.isPending}
						<p role="status" class="flex items-center gap-2 text-sm text-muted-foreground">
							<Spinner /> Loading latest report
						</p>
					{:else if investigationController.reportQuery.isError && investigationController.reportQuery.error?.status !== 404}
						<p role="alert" class="text-sm text-muted-foreground">
							The latest report could not be loaded.
						</p>
						<Button
							variant="outline"
							size="sm"
							onclick={() => investigationController.reportQuery.refetch()}
						>
							Retry
						</Button>
					{:else}
						<p class="text-sm text-muted-foreground">No report yet.</p>
					{/if}

					<Button variant="link" href={sitCtrl.investigationHref} class="self-start">
						View investigation
					</Button>
				</Card.Content>
			</Card.Root>

			<section aria-labelledby="evidence-title" class="flex flex-col gap-3">
				<div class="flex flex-wrap items-baseline justify-between gap-3">
					<h2 id="evidence-title" class="text-lg font-semibold">Evidence</h2>
					<p class="text-xs text-muted-foreground tabular-nums">
						{attrs.signalCount}
						{attrs.signalCount === 1 ? "source" : "sources"}
					</p>
				</div>

				{#each controller.observations as group (group.id)}
					<Collapsible.Root
						open={controller.groupOpen(group.id)}
						onOpenChange={(open) => controller.setGroupOpen(group.id, open)}
						class="rounded-lg border bg-card"
					>
						<Collapsible.Trigger
							class="group flex w-full items-center gap-3 p-4 text-left text-sm font-medium hover:bg-accent focus-visible:outline-2 focus-visible:outline-ring"
						>
							{#snippet child({ props })}
								<div {...props}>
									<RiFileListLine class="size-6" aria-hidden="true" />
									<span class="min-w-0 flex-1 wrap-anywhere">{group.title}</span>
									<span class="shrink-0 text-xs text-muted-foreground tabular-nums">
										{group.records.length}
										{group.records.length === 1 ? "source" : "sources"}
									</span>
									<RiArrowDownSLine
										class="size-6 group-data-[state=open]:rotate-180"
										aria-hidden="true"
									/>
								</div>
							{/snippet}
						</Collapsible.Trigger>
						<Collapsible.Content>
							{#if group.body}
								<div class="flex flex-col gap-1 border-t px-4 py-3">
									<h3 class="text-xs font-medium text-muted-foreground">Observation</h3>
									<p class="max-w-[75ch] whitespace-pre-wrap text-sm wrap-anywhere">
										{group.body}
									</p>
								</div>
							{/if}
							<ul>
								{#each group.records as record (record.key)}
									<li class="flex flex-col gap-3 border-t p-4 sm:flex-row sm:items-center">
										<div class="flex min-w-0 flex-1 flex-col gap-1">
											<p class="text-[15px] font-medium wrap-anywhere">
												{record.title}
											</p>
											<p
												class="line-clamp-2 whitespace-pre-wrap text-sm text-muted-foreground wrap-anywhere"
											>
												{record.content || "Source content unavailable."}
											</p>
										</div>
										<div
											class="flex min-w-0 flex-col gap-1 text-xs text-muted-foreground sm:w-52"
										>
											<p class="wrap-anywhere">{record.type} · {record.source}</p>
											<time datetime={record.time.iso} class="tabular-nums">
												{record.timeLabel}
												{record.time.label}
											</time>
										</div>
										<Button
											variant="ghost"
											size="sm"
											class="self-start sm:self-center"
											aria-label={`Inspect ${record.title}`}
											onclick={(event) =>
												controller.openSource(
													{
														kind: "direct",
														record,
														observationGroupTitle: group.title,
													},
													event.currentTarget
												)}
										>
											Inspect
										</Button>
									</li>
								{:else}
									<li>
										<Empty.Root>
											<Empty.Header>
												<Empty.Title>No source records</Empty.Title>
											</Empty.Header>
										</Empty.Root>
									</li>
								{/each}
							</ul>
						</Collapsible.Content>
					</Collapsible.Root>
				{:else}
					<Empty.Root>
						<Empty.Header>
							<Empty.Title>No grouped evidence recorded</Empty.Title>
						</Empty.Header>
					</Empty.Root>
				{/each}
			</section>
		</div>
	</div>
{/if}

<SituationSourceSheet target={controller.sourceTarget} onClose={controller.closeSource} />
