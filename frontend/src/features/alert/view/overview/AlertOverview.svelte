<script lang="ts">
	import { useAlertViewState } from "$features/alert";
	import MetricCard from "$src/components/viz/MetricCard.svelte";
	import {
		mdiClipboardText,
		mdiLineScan,
		mdiMoonWaxingCrescent,
		mdiPhoneAlert,
	} from "@mdi/js";

	const viewState = useAlertViewState();

	const attrs = $derived(viewState.alert?.attributes);
	const metrics = $derived(viewState.metrics);
	const notAccurateFbs = $derived(
		!!metrics ? metrics.feedbacks - metrics.accurate - metrics.accurateUnknown : 0
	);
	const accuracy = $derived(
		!!metrics ? `${metrics.accurate}/${notAccurateFbs}/${metrics.accurateUnknown}` : ""
	);
</script>

{#snippet infoCard(title: string, text: string)}
	<div class="flex flex-col gap-2 border p-2 h-fit">
		<span class="uppercase font-semibold text-surface-content/90">{title}</span>

		<span>{text}</span>
	</div>
{/snippet}

<div class="flex gap-2">
	<div class="flex flex-col gap-2">
		{@render infoCard("Description", attrs?.description || "")}
		{@render infoCard("Owners", "")}
		{@render infoCard("Definition", "")}
		{@render infoCard("Source", "")}
	</div>

	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Metrics</span>

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
