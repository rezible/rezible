<script lang="ts">
	import { startOfMonth, endOfMonth } from "date-fns";
	import { Calendar, Chart, Tooltip, Svg, BarChart, Axis, Bars, LinearGradient } from "layerchart";
	import { format as lsFormat, PeriodType } from "@layerstack/utils";
	import { Button, Header } from "svelte-ux";
	import { mdiDotsVertical } from "@mdi/js";
	import { scaleBand } from "d3-scale";
	import { fakeDateSeries, createDateSeries } from "../../lib/genData";

	const now = new Date();
	const firstDayOfMonth = startOfMonth(now);
	const lastDayOfMonth = endOfMonth(now);

	const dateSeriesData = createDateSeries({
		count: 10,
		min: 20,
		max: 100,
		value: "integer",
		keys: ["value", "baseline"],
	});
</script>

<div class="h-fit flex flex-col gap-2 border rounded px-4 py-2">
	<Header>
		<span slot="title" class="text-lg">
			Incidents <span class="opacity-80">(past 30 days)</span>
		</span>
		<svelte:fragment slot="actions">
			<Button icon={mdiDotsVertical} iconOnly />
		</svelte:fragment>
	</Header>

	<div class="h-[300px] p-4 border rounded bg-surface-100">
		<!--BarChart
			data={fakeDateSeries}
			x="date"
			y="value"
			renderContext="svg"
			debug={false}
			c="value"
			cRange={[
				"oklch(var(--color-success))",
				"oklch(var(--color-warning))",
				"oklch(var(--color-danger))",
			]}
		/-->
		<BarChart data={fakeDateSeries} x="date" y="value" renderContext="canvas" debug={false} tooltip={false}>
			<svelte:fragment slot="marks" let:series let:getBarsProps>
				{#each series as s, i (s.key)}
					<LinearGradient
						class="from-danger-500 to-success-400"
						vertical
						units="userSpaceOnUse"
						let:gradient
					>
						<Bars {...getBarsProps(s, i)} fill={gradient} />
					</LinearGradient>
				{/each}
			</svelte:fragment>
		</BarChart>
	</div>

	<!--div class="h-64">
		<BarChart
			data={dateData}
			x="date"
			y="value"
			grid={{ x: true, y: true }}
			tooltip={false}
			cRange={[
				"hsl(var(--color-danger))",
				"hsl(var(--color-warning))",
				"hsl(var(--color-success))",
				"hsl(var(--color-info))",
			]}
		/>
	</div-->
</div>
