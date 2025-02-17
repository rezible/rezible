<script lang="ts">
	import { format, PeriodType } from "@layerstack/utils";
	import { startOfMonth, endOfMonth } from "date-fns";
	import { scaleThreshold } from "d3-scale";

	import { Calendar, Chart, Tooltip, Svg, BarChart } from "layerchart";
	import { createDateSeries } from "$features/reports/lib/genData";
	import { Button, Header } from "svelte-ux";
	import { mdiDotsVertical } from "@mdi/js";

	const now = new Date();
	const firstDayOfMonth = startOfMonth(now);
	const lastDayOfMonth = endOfMonth(now);

	const dateData = createDateSeries({ count: 31, min: 0, max: 6, value: "integer" }).map((d) => {
		return {
			...d,
			value: Math.random() > 0.4 ? d.value : null,
		};
	});
</script>

<div class="h-64 flex flex-col gap-2 border rounded px-4 py-2">
	<Header>
		<span slot="title" class="text-lg">
			Incidents <span class="opacity-80">(past 30 days)</span>
		</span>
		<svelte:fragment slot="actions">
			<Button icon={mdiDotsVertical} iconOnly />
		</svelte:fragment>
	</Header>

	<div class="flex-1">
		<BarChart
		data={dateData}
		x="date"
		y="value"
		grid={{ x: true }}
		tooltip={false}
		cRange={[
			"hsl(var(--color-danger))",
			"hsl(var(--color-warning))",
			"hsl(var(--color-success))",
			"hsl(var(--color-info))",
		]}
		/>
	</div>
</div>