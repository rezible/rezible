<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Input } from "$components/ui/input";
	import { Spinner } from "$components/ui/spinner";
	import * as Field from "$components/ui/field";
	import * as Alert from "$components/ui/alert";
	import * as Empty from "$components/ui/empty";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import SituationInvestigationList from "../SituationInvestigationList.svelte";
	import RiInformationLine from "remixicon-svelte/icons/information-line";
	import { initSituationInvestigationsController } from "./controller.svelte";
	import { timestamp } from "../model";

	const controller = initSituationInvestigationsController();
	const attrs = $derived(controller.attrs);
	const situation = controller.situationController;
</script>

{#if attrs}
	<div class="min-h-0 min-w-0 flex-1 overflow-y-auto p-4">
		<div class="mx-auto flex max-w-7xl flex-col gap-6">
			<h1 class="text-[28px] leading-9 font-semibold">Investigations</h1>
			<form
				class="max-w-2xl"
				onsubmit={(e) => {
					e.preventDefault();
					void controller.runInvestigation();
				}}
			>
				<Field.FieldGroup>
					<Field.Field>
						<Field.FieldLabel for="investigation-query">Question (optional)</Field.FieldLabel>
						<div class="flex flex-wrap gap-2">
							<Input
								id="investigation-query"
								class="min-w-40 flex-1"
								bind:value={controller.form.query}
								disabled={controller.form.pending}
								placeholder="What would you like to understand?"
							/>
							<Button type="submit" disabled={controller.form.pending}>
								{#if controller.form.pending}
									<Spinner data-icon="inline-start" />
								{/if}
								{controller.form.pending ? "Requesting investigation…" : "Run investigation"}
							</Button>
						</div>
					</Field.Field>
				</Field.FieldGroup>
			</form>

			{#if controller.startMutation.error}
				<Alert.Root variant="destructive">
					<Alert.Title>Could not request investigation</Alert.Title>
					<Alert.Description
						>{controller.startMutation.error.detail ||
							controller.startMutation.error.title}</Alert.Description
					>
				</Alert.Root>
			{/if}

			<div class="grid min-w-0 items-start gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
				<article class="flex min-w-0 flex-col gap-5 rounded-lg border bg-card p-4 md:p-6">
					{#if situation.selectedInvestigationId}
						<LoadingQueryWrapper query={situation.selectedInvestigationQuery} feedbackOnly />
						{#if situation.selectedInvestigation && !situation.selectedInvestigationBelongs}
							<Alert.Root>
								<Alert.Title>Investigation unavailable for this Situation</Alert.Title>
								<Alert.Description
									>Select an investigation from this Situation's list.</Alert.Description
								>
							</Alert.Root>
						{:else if situation.selectedInvestigation}
							{@const selAttrs = situation.selectedInvestigation.attributes}
							<header class="flex flex-col gap-2">
								<h2 class="text-lg font-semibold wrap-anywhere">
									{selAttrs.query || "Investigation"}
								</h2>
								<time
									class="text-xs text-muted-foreground tabular-nums"
									datetime={timestamp(selAttrs.updatedAt).iso}
									>Updated {timestamp(selAttrs.updatedAt).label}</time
								>
							</header>
							{#if controller.selectedChanged}
								<Alert.Root role="note"
									><RiInformationLine /><Alert.Title
										><span class="text-status-warning-foreground"
											>Evidence has changed since this report.</span
										></Alert.Title
									></Alert.Root
								>
							{/if}
							{#if controller.selectedReport}
								<p
									class="max-w-[75ch] whitespace-pre-wrap text-[15px] leading-6 wrap-anywhere"
								>
									{controller.selectedReport.text || "Report text unavailable."}
								</p>
								{#each controller.reportSections as section (section.title)}
									<section class="flex max-w-[75ch] flex-col gap-2">
										<h3 class="font-semibold">{section.title}</h3>
										{#if section.items.length === 1}
											<p
												class="whitespace-pre-wrap text-[15px] leading-6 wrap-anywhere"
											>
												{section.items[0]}
											</p>
										{:else}
											<ul
												class="flex list-disc flex-col gap-2 pl-5 text-[15px] leading-6"
											>
												{#each section.items as item, index (index)}<li
														class="whitespace-pre-wrap wrap-anywhere"
													>
														{item}
													</li>{/each}
											</ul>
										{/if}
									</section>
								{/each}
							{:else}
								<Empty.Root>
									<Empty.Header>
										<Empty.Title>No report available yet</Empty.Title>
									</Empty.Header>
								</Empty.Root>
							{/if}
						{/if}
					{:else}
						<Empty.Root>
							<Empty.Header>
								<Empty.Title>No investigation yet</Empty.Title>
								<Empty.Description
									>Run an investigation using the form above.</Empty.Description
								>
							</Empty.Header>
						</Empty.Root>
					{/if}
				</article>

				<SituationInvestigationList />
			</div>
		</div>
	</div>
{/if}
