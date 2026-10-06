<script lang="ts">
	import * as Collapsible from "$components/ui/collapsible";
	import { Button } from "$components/ui/button";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import { resolve } from "$app/paths";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { sourceIcon, sourceMeta } from "$features/situations/lib/model";
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";
	import RiContractUpDownLine from "remixicon-svelte/icons/contract-up-down-line";
	import RiExpandUpDownLine from "remixicon-svelte/icons/expand-up-down-line";
	import RiFileList3Line from "remixicon-svelte/icons/file-list-3-line";
	import { useSituationOverviewController } from "./controller.svelte";

	const controller = useSituationOverviewController();
	const observations = $derived(controller.observations);
	const sourceLabel = $derived(controller.sourceCount === 1 ? "source" : "sources");
</script>

{#snippet headingActions()}
	<span class="text-xs text-muted-foreground tabular-nums">{controller.sourceCount} {sourceLabel}</span>
	{#if observations.length > 1}
		<Button variant="ghost" size="sm" onclick={controller.expandAll}>
			<RiExpandUpDownLine data-icon="inline-start" />
			Expand all
		</Button>
		<Button variant="ghost" size="sm" onclick={controller.collapseAll}>
			<RiContractUpDownLine data-icon="inline-start" />
			Collapse all
		</Button>
	{/if}
{/snippet}

<section aria-labelledby="observations-title" class="flex flex-col gap-3">
	<SectionHeading
		id="observations-title"
		title="Observations"
		count={observations.length}
		actions={headingActions}
	/>

	{#each observations as group (group.id)}
		<Collapsible.Root
			open={controller.groupOpen(group.id)}
			onOpenChange={(open) => controller.setGroupOpen(group.id, open)}
			class="rounded-lg border bg-card"
		>
			<Collapsible.Trigger
				class="group flex min-h-14 w-full items-center gap-3 rounded-lg px-4 text-left hover:bg-accent focus-visible:outline-2 focus-visible:outline-ring data-[state=open]:rounded-b-none"
			>
				<RiFileList3Line class="size-5 shrink-0 text-muted-foreground" aria-hidden="true" />
				<span class="min-w-0 flex-1 text-[15px] leading-[22px] font-medium wrap-anywhere">
					{group.title}
				</span>
				<span class="shrink-0 text-xs text-muted-foreground tabular-nums">
					{group.records.length}
					{group.records.length === 1 ? "source" : "sources"}
				</span>
				<RiArrowDownSLine
					class="size-5 shrink-0 transition-transform duration-150 group-data-[state=open]:rotate-180 motion-reduce:transition-none"
					aria-hidden="true"
				/>
			</Collapsible.Trigger>
			<Collapsible.Content>
				{#if group.body}
					<div class="flex flex-col gap-1 border-t px-4 py-3">
						<h3 class="region-label">Observation</h3>
						<p class="max-w-[72ch] text-sm leading-[22px] whitespace-pre-wrap wrap-anywhere">
							{group.body}
						</p>
					</div>
				{/if}
				{#if group.records.length}
					<ul class="divide-y border-t">
						{#each group.records as record (record.key)}
							{@const Icon = sourceIcon(record)}
							<li
								class="grid min-h-[60px] grid-cols-[20px_minmax(0,1fr)] items-start gap-x-3 gap-y-2 px-4 py-3 sm:grid-cols-[20px_minmax(0,1fr)_12rem_auto] sm:items-center"
							>
								<Icon
									class="mt-0.5 size-5 text-muted-foreground sm:mt-0"
									aria-hidden="true"
								/>
								<div class="flex min-w-0 flex-col gap-1">
									<p
										class="flex flex-wrap items-center gap-x-2 gap-y-1 text-[15px] leading-[22px] font-medium wrap-anywhere"
									>
										{record.title}
										{#if record.attention && record.definitionId}
											<a
												class="rounded-sm hover:underline focus-visible:outline-2 focus-visible:outline-ring"
												href={resolve("/signals/[id]/[[view=signalView]]", {
													id: record.definitionId,
												})}
											>
												<StatusBadge status={record.attention} variant="inline" />
											</a>
										{/if}
									</p>
									<p
										class="line-clamp-2 text-sm whitespace-pre-wrap text-muted-foreground wrap-anywhere"
									>
										{record.content || "No description from the source."}
									</p>
								</div>
								<div
									class="col-start-2 flex min-w-0 flex-col gap-0.5 text-xs text-muted-foreground sm:col-start-auto"
								>
									<span class="wrap-anywhere">{sourceMeta(record)}</span>
									<span>
										{record.timeLabel}
										<Timestamp value={record.time.iso} />
									</span>
								</div>
								<Button
									variant="ghost"
									size="sm"
									class="col-start-2 justify-self-start sm:col-start-auto"
									aria-label={`Inspect ${record.title}`}
									onclick={(event) =>
										controller.inspectSource(record, group.title, event.currentTarget)}
								>
									Inspect
								</Button>
							</li>
						{/each}
					</ul>
				{:else}
					<p class="border-t px-4 py-3 text-sm text-muted-foreground">No sources in this group.</p>
				{/if}
			</Collapsible.Content>
		</Collapsible.Root>
	{:else}
		<p class="text-sm text-muted-foreground">No observations have been grouped for this situation.</p>
	{/each}
</section>
