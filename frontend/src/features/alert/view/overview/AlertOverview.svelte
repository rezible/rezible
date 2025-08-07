<script lang="ts">
	import { useAlertViewState } from "$features/alert";
	import { AnnotationDialogState, setAnnotationDialogState } from "$src/components/oncall-events/annotation-dialog/dialogState.svelte";
	import EventRow from "$src/components/oncall-events/EventRow.svelte";
	import PaginatedListBox from "$src/components/paginated-listbox/PaginatedListBox.svelte";
	import MetricCard from "$src/components/viz/MetricCard.svelte";
	import {
		mdiClipboardText,
		mdiLineScan,
		mdiMoonWaxingCrescent,
		mdiPhoneAlert,
	} from "@mdi/js";

	const viewState = useAlertViewState();

	const attrs = $derived(viewState.alert?.attributes);
	const playbooks = $derived(attrs?.linkedPlaybooks);
	const metrics = $derived(viewState.metrics);
	const events = $derived(viewState.events);
	const notAccurateFbs = $derived(
		!!metrics ? metrics.feedbacks - metrics.accurate - metrics.accurateUnknown : 0
	);
	const accuracy = $derived(
		!!metrics ? `${metrics.accurate}/${notAccurateFbs}/${metrics.accurateUnknown}` : ""
	);

	setAnnotationDialogState(new AnnotationDialogState({}));
</script>

<div class="flex gap-2">
	<div class="flex flex-col gap-2">
		<div class="flex flex-col gap-2 w-full border p-2">
			<span class="uppercase font-semibold text-surface-content/90">Description</span>

			<span>{attrs?.description ?? ""}</span>
		</div>

		<div class="flex flex-col gap-2 w-fit border p-2">
			<span class="uppercase font-semibold text-surface-content/90">Metrics</span>
			<span>from {viewState.metricsFrom.toAbsoluteString()}</span>
			<span>to {viewState.metricsTo.toAbsoluteString()}</span>

			{#if metrics}
				<div class="flex">
					<div class="">
						<h1>Events</h1>
						<MetricCard title="Trigger Events" icon={mdiLineScan} metric={metrics.triggers} />
						<MetricCard title="Interrupts" icon={mdiPhoneAlert} metric={metrics.interrupts} />
						<MetricCard
							title="Night Interrupts"
							icon={mdiMoonWaxingCrescent}
							metric={metrics.nightInterrupts}
						/>
					</div>
					<div class="">
						<h1>Feedback</h1>
						<MetricCard title="Feedback Given" icon={mdiClipboardText} metric={metrics.feedbacks} />
						<MetricCard
							title="Actionable"
							icon={mdiClipboardText}
							metric={metrics.actionable / metrics.feedbacks}
							format="percentage"
						/>
						<MetricCard title="Accurate (Yes/No/Unknown)" icon={mdiClipboardText} metric={accuracy} />
						<MetricCard
							title="Documentation Available"
							icon={mdiClipboardText}
							metric={metrics.docsAvailable / metrics.feedbacks}
						/>
					</div>
				</div>
			{/if}
		</div>
	</div>

	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Linked Playbooks</span>

		{#if playbooks}
			{#each playbooks as pb}
				<a class="border p-1" href="/playbooks/{pb.id}">{pb.attributes.title}</a>
			{:else}
				<span>no playbooks linked</span>
			{/each}
		{/if}
	</div>

	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Events</span>

		{#if events}
			<PaginatedListBox {...viewState.eventsQueryPagination.paginationProps} title="">
				{#each events as ev}
					<EventRow event={ev} />
				{:else}
					<span>no events</span>
				{/each}
			</PaginatedListBox>
		{/if}
	</div>
</div>
