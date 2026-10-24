<script lang="ts">
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import RiCloseLine from "remixicon-svelte/icons/close-line";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import { Button } from "$components/ui/button";
	import * as Dialog from "$components/ui/dialog";
	import { Spinner } from "$components/ui/spinner";
	import SituationCombobox from "$features/situations/components/situation-combobox/SituationCombobox.svelte";
	import { situationStageStatus } from "$features/situations/lib/status";

	import { useIncidentCreateDialog } from "./controller.svelte";
	import IncidentDetailsFields from "./IncidentDetailsFields.svelte";

	const controller = useIncidentCreateDialog();
	const chosenIds = $derived(controller.form.situations.map((situation) => situation.id));
</script>

<Dialog.Root bind:open={() => controller.open, controller.setOpen}>
	<Dialog.Content class="max-h-[calc(100dvh-2rem)] overflow-hidden p-0 sm:max-w-[640px]" showCloseButton>
		<form
			class="flex max-h-[calc(100dvh-2rem)] flex-col"
			onsubmit={(event) => {
				event.preventDefault();
				controller.submit();
			}}
		>
			<div class="border-b px-4 py-3">
				<Dialog.Header>
					<Dialog.Title>Declare incident</Dialog.Title>
					<Dialog.Description>
						Start an incident response. Only a title is needed.
					</Dialog.Description>
				</Dialog.Header>
			</div>

			<div class="flex flex-1 flex-col gap-4 overflow-y-auto p-4">
				{#if controller.metadataQuery.isLoading}
					<div class="grid h-52 w-full place-items-center">
						<Spinner />
					</div>
				{:else}
					<InlineAlert error={controller.metadataError ?? undefined} dismissable={false} />
					<IncidentDetailsFields
						form={controller.form}
						metadata={controller.metadata}
						titleError={controller.titleError}
						isAdmin={controller.isAdmin}
						idPrefix="declare-incident"
					/>

					<div class="flex flex-col gap-2">
						<span class="text-sm font-medium">Linked situations</span>
						{#if controller.form.situations.length}
							<ul class="flex flex-col gap-2">
								{#each controller.form.situations as situation (situation.id)}
									{@const existing = controller.existingIncidentFor(situation)}
									<li class="flex flex-col gap-1">
										<span
											class="inline-flex max-w-full items-center gap-2 self-start rounded-full border bg-card py-1 pr-1 pl-3 text-sm"
										>
											<span class="truncate">{situation.title}</span>
											<StatusBadge
												status={situationStageStatus(situation.stage)}
												variant="inline"
											/>
											<Button
												variant="ghost"
												size="icon-sm"
												class="size-6"
												aria-label="Remove {situation.title}"
												onclick={() => controller.removeSituation(situation.id)}
											>
												<RiCloseLine />
											</Button>
										</span>
										{#if existing}
											<p class="text-xs text-muted-foreground">
												Already part of {existing.title} ·
												<a class="underline underline-offset-4" href={existing.href}>
													Open it
												</a>
											</p>
										{/if}
									</li>
								{/each}
							</ul>
						{/if}
						<SituationCombobox linkedIds={chosenIds} onSelect={controller.addSituation}>
							{#snippet trigger(props)}
								<Button {...props} variant="outline" size="sm" class="self-start">
									<RiAddLine />
									Add situation
								</Button>
							{/snippet}
						</SituationCombobox>
					</div>
				{/if}
			</div>

			<div class="flex flex-col gap-3 border-t px-4 py-3">
				<InlineAlert error={controller.error} dismissable={false} />
				<Dialog.Footer>
					<Button type="button" variant="ghost" onclick={() => controller.setOpen(false)}>
						Cancel
					</Button>
					<Button type="submit" disabled={controller.isPending}>
						{#if controller.isPending}
							<Spinner data-icon="inline-start" />
						{/if}
						Declare incident
					</Button>
				</Dialog.Footer>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
