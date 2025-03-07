<script lang="ts">
	import { Header, Icon } from "svelte-ux";
	import { mdiChevronRight, mdiAlarmLight, mdiFire, mdiSleepOff } from "@mdi/js";
	import type { OncallShift } from "$src/lib/api";
	import { Highlight, Chart, Tooltip, Svg, BarChart, Axis, Bars, LinearGradient } from "layerchart";
	import { format as lsFormat, PeriodType } from "@layerstack/utils";
	import { scaleBand } from "d3-scale";
	import { parseAbsoluteToLocal } from "@internationalized/date";

	type Props = {
		shift: OncallShift;
	};
	const { shift }: Props = $props();

	const hourData = (hour: number, value: number, baseline: number) => {
		const period = hour < 12 ? "am" : "pm";
		const hour12 = hour % 12 || 12;
		// format={(d) => lsFormat(d, PeriodType.TimeOnly, { variant: "short", custom: { withOrdinal: false } })}
		return { date: `${hour12} ${period}`, value, baseline };
	}
	const alertData = $derived([
		hourData(0, 2, 0),
		hourData(1, 0, 0),
		hourData(2, 0, 0),
		hourData(3, 0, 0),
		hourData(4, 0, 0),
		hourData(5, 0, 0),
		hourData(6, 0, 0),
		hourData(7, 0, 0),
		hourData(8, 0, 0),
		hourData(9, 1, 1),
		hourData(10, 0, 0),
		hourData(11, 0, 0),
		hourData(12, 2, 4),
		hourData(13, 0, 0),
		hourData(14, 0, 0),
		hourData(15, 0, 0),
		hourData(16, 0, 0),
		hourData(17, 0, 0),
		hourData(18, 0, 0),
		hourData(19, 0, 0),
		hourData(20, 0, 0),
		hourData(21, 0, 0),
		hourData(22, 0, 0),
		hourData(23, 0, 0),
	]);
</script>

<div class="flex flex-col gap-2 flex-1 min-h-0 max-h-full overflow-y-auto border rounded-lg p-2">
	<Header title="Stats" />

	<div class="flex flex-row gap-4">
		<div class="flex items-center gap-4 bg-info-900/80 h-full p-2 px-6 rounded-xl w-fit">
			<div class="">
				<Icon data={mdiAlarmLight} />
			</div>
			<div class="flex flex-col">
				<span class="text-lg">5 Alerts</span>
				<span class="text-surface-content/75">Normal</span>
			</div>
		</div>
		<div class="flex items-center gap-4 bg-accent-900/80 h-full p-3 rounded-xl w-fit">
			<div class="">
				<Icon data={mdiSleepOff} />
			</div>
			<div class="flex flex-col">
				<span class="text-lg">1 Alert at Night</span>
				<span class="text-surface-content/75">Above Average</span>
			</div>
		</div>
		<div class="flex items-center gap-4 bg-warning-900/80 h-full p-3 rounded-xl w-fit">
			<div class="">
				<Icon data={mdiFire} />
			</div>
			<div class="flex flex-col">
				<span class="text-lg">2 Incidents</span>
				<span class="text-surface-content/75">Above Average</span>
			</div>
		</div>
	</div>

	<Header title="Disruption by Hour" />

	<div class="h-[500px] p-4 border rounded">
		<Chart
			data={alertData}
			x="date"
			yDomain={[0, null]}
			yNice
			y={["value", "baseline"]}
			xScale={scaleBand().padding(0.2)}
			padding={{ left: 6, bottom: 6 }}
			tooltip={{ mode: "band" }}
		>
			<Svg>
				<Axis placement="left" grid rule format={d => lsFormat(d, "integer")} />
				<Axis placement="bottom" rule />
				<Bars y="baseline" strokeWidth={1} class="fill-surface-content/20" />
				<Bars y="value" strokeWidth={1} insets={{ x: 8 }} class="fill-primary" />
				<Highlight area />
			</Svg>
			<Tooltip.Root let:data>
				<Tooltip.Header
					>{lsFormat(data.date, PeriodType.Custom, {
						custom: "eee, MMMM do",
					})}</Tooltip.Header
				>
				<Tooltip.List>
					<Tooltip.Item label="value" value={data.value} />
					<Tooltip.Item label="baseline" value={data.baseline} />
				</Tooltip.List>
			</Tooltip.Root>
		</Chart>
	</div>
</div>
