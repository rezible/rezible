<script lang="ts">
	import { Button } from "$components/ui/button";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { initAlertInvestigationController } from "./controller.svelte";

	const controller = initAlertInvestigationController();
</script>

<LoadingQueryWrapper query={controller.casesQuery}>
	{#snippet view()}
		<div class="grid min-w-0 gap-4 p-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
			<section class="flex min-w-0 flex-col gap-4">
				<div class="flex flex-wrap items-center justify-between gap-3 border border-border p-3">
					<div class="min-w-0">
						<h2 class="text-base font-semibold text-foreground">Alert investigation</h2>
						<p class="truncate text-sm text-muted-foreground">{controller.title}</p>
					</div>
					<Button onclick={controller.runInvestigation} disabled={controller.requestDisabled}>
						{#if controller.requestPending}
							<Spinner />
							Requesting
						{:else}
							Run Investigation
						{/if}
					</Button>
				</div>

				{#if controller.resultPayload}
					<div class="grid gap-3 border border-border p-3">
						<div class="flex flex-wrap items-center justify-between gap-2">
							<h3 class="text-sm font-semibold text-foreground">Agent findings</h3>
							<span class="text-xs uppercase text-muted-foreground"
								>{controller.statusLabel(controller.latestRun)}</span
							>
						</div>
						{#if controller.resultSummary}
							<p class="text-sm leading-6">{controller.resultSummary}</p>
						{/if}
						{#if controller.likelyCause}
							<div class="grid gap-1">
								<span class="text-xs font-semibold uppercase text-muted-foreground"
									>Likely cause</span
								>
								<p class="text-sm leading-6">{controller.likelyCause}</p>
							</div>
						{/if}
						{#if controller.findingSuggestedChecks.length > 0}
							<div class="grid gap-1">
								<span class="text-xs font-semibold uppercase text-muted-foreground"
									>Suggested checks</span
								>
								<ul class="list-disc space-y-1 pl-5 text-sm">
									{#each controller.findingSuggestedChecks as check (check)}
										<li>{check}</li>
									{/each}
								</ul>
							</div>
						{/if}
						{#if controller.recommendedNext}
							<div class="grid gap-1">
								<span class="text-xs font-semibold uppercase text-muted-foreground"
									>Recommended next step</span
								>
								<p class="text-sm leading-6">{controller.recommendedNext}</p>
							</div>
						{/if}
					</div>
				{:else}
					<div class="border border-border p-3 text-sm text-muted-foreground">
						{#if controller.latestRun}
							Agent run status: {controller.statusLabel(controller.latestRun)}.
						{:else}
							Run the investigation to synthesize findings and post to the alert Slack thread
							when one exists.
						{/if}
					</div>
				{/if}

				<div class="grid gap-3 md:grid-cols-2">
					<section class="border border-border p-3">
						<h3 class="mb-2 text-sm font-semibold text-foreground">Likely subjects</h3>
						{#if controller.subjectItems.length > 0}
							<div class="grid gap-2">
								{#each controller.subjectItems as subject (subject.id)}
									<div class="grid gap-1 border-l border-border pl-3">
										<span class="text-sm font-medium">{subject.displayName}</span>
										<span class="text-xs uppercase text-muted-foreground"
											>{subject.kind} · {subject.confidence}</span
										>
										{#if subject.reason}
											<p class="text-xs leading-5 text-muted-foreground">
												{subject.reason}
											</p>
										{/if}
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-sm text-muted-foreground">No linked subjects found.</p>
						{/if}
					</section>

					<section class="border border-border p-3">
						<h3 class="mb-2 text-sm font-semibold text-foreground">Recent signals</h3>
						{#if controller.signalItems.length > 0}
							<div class="grid gap-2">
								{#each controller.signalItems as signal (signal.id)}
									<div class="grid gap-1 border-l border-border pl-3">
										<span class="text-sm">{signal.summary}</span>
										<p class="text-xs leading-5 text-muted-foreground">
											{signal.source} · {signal.kind}
										</p>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-sm text-muted-foreground">No recent evidence found.</p>
						{/if}
					</section>
				</div>
			</section>

			<aside class="flex min-w-0 flex-col gap-4">
				<section class="border border-border p-3">
					<h3 class="mb-2 text-sm font-semibold text-foreground">Investigation guide</h3>
					{#if controller.suggestedChecks.length > 0}
						<ul class="list-disc space-y-1 pl-5 text-sm">
							{#each controller.suggestedChecks as check (check)}
								<li>{check}</li>
							{/each}
						</ul>
					{:else}
						<p class="text-sm text-muted-foreground">No suggested checks available.</p>
					{/if}
				</section>

				<section class="border border-border p-3">
					<h3 class="mb-2 text-sm font-semibold text-foreground">Related knowledge</h3>
					{#if controller.neighborItems.length > 0}
						<div class="grid gap-2 text-sm">
							{#each controller.neighborItems as neighbor (neighbor.id)}
								<div class="grid gap-1">
									<span class="font-medium">{neighbor.name}</span>
									<span class="text-xs text-muted-foreground">
										{neighbor.relatedEntityKind} · {neighbor.kind}
									</span>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-muted-foreground">
							No neighboring knowledge graph items found.
						</p>
					{/if}
				</section>

				{#if controller.guideItems.length > 0}
					<section class="border border-border p-3">
						<h3 class="mb-2 text-sm font-semibold text-foreground">
							Playbooks and prior incidents
						</h3>
						<div class="grid gap-2 text-sm">
							{#each controller.guideItems as guide (guide.id)}
								<div class="grid gap-1">
									<span class="font-medium">{guide.title}</span>
									<p class="text-xs leading-5 text-muted-foreground">{guide.summary}</p>
								</div>
							{/each}
						</div>
					</section>
				{/if}
			</aside>
		</div>
	{/snippet}
</LoadingQueryWrapper>
