<script lang="ts">
	import type { BacklogItem } from "../types";
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { rosterIdCtx } from "$features/oncall/views/roster/context";
	import { Header } from "svelte-ux";
	import { Axis, Bars, Chart, Points, Svg, Tooltip, Highlight } from "layerchart";
	import { scaleBand, scaleOrdinal } from "d3-scale";
	import { PeriodType, formatDate } from "@layerstack/utils";
	import { getLocalTimeZone, now } from "@internationalized/date";
	
	const rosterId = rosterIdCtx.get();

	const mockBacklogItems: BacklogItem[] = [
		// TODO: fill this with some items
	];

	const getRosterBacklogOptions = (options: {path: {id: string}}) => {
		return queryOptions({
			queryFn: async ({ queryKey, signal }) => {
				return {data: mockBacklogItems};
			},
			queryKey: ["getUserOncallDetailsOptions", options.path.id]
		});
	};

	const backlogQuery = createQuery(() => getRosterBacklogOptions({path: {id: rosterId}}));

	const today = now(getLocalTimeZone());

	type TicketBurn = {
		"date": Date,
		"start": number,
		"high": number,
		"low": number,
		"end": number,
	}
	const burndownData: TicketBurn[] = [
		{
			"date": today.subtract({weeks: 4}).toDate(),
			"start": 3,
			"high": 8,
			"low": 3,
			"end": 4,
		},
		{
			"date": today.subtract({weeks: 3}).toDate(),
			"start": 4,
			"high": 6,
			"low": 3,
			"end": 3,
		},
		{
			"date": today.subtract({weeks: 2}).toDate(),
			"start": 3,
			"high": 11,
			"low": 3,
			"end": 6,
		},
		{
			"date": today.subtract({weeks: 1}).toDate(),
			"start": 6,
			"high": 14,
			"low": 6,
			"end": 10,
		},
		{
			"date": today.toDate(),
			"start": 10,
			"high": 11,
			"low": 6,
			"end": 6,
		},
	]
</script>

<div class="grid grid-cols-2 gap-2 h-full">
	<div class="flex flex-col">
		<Header title="Tickets" class="text-lg font-medium" />

		<div class="flex-1 flex flex-col gap-1 overflow-y-auto border">
			<LoadingQueryWrapper query={backlogQuery}>
				{#snippet view(items: BacklogItem[])}
					{#each items as item, i}
						<!-- TODO: create this list -->
						<span>{JSON.stringify(item)}</span>
					{:else}
						<div class="text-surface-600 italic p-2">Backlog empty!</div>
					{/each}
				{/snippet}
			</LoadingQueryWrapper>
		</div>
	</div>

	<div class="">
		<Header title="Ticket Burndown" class="text-lg font-medium" />

		<div class="h-[300px] p-4 border rounded">
			<Chart
			  data={burndownData}
			  x="date"
			  xScale={scaleBand()}
			  y={["high", "low"]}
			  yNice
			  c={(d: TicketBurn) => (d.end === d.start ? "same" : (d.end < d.start ? "desc" : "asc"))}
			  cScale={scaleOrdinal()}
			  cDomain={["desc", "same", "asc"]}
			  cRange={["rgb(90, 190, 90)", "#333333", "rgb(190, 90, 90)"]}
			  tooltip={{ mode: "bisect-x" }}
			>
			  <Svg>
				<Axis placement="left" grid rule />
				<Axis placement="bottom" rule format={(d: string) => ""} />
				<Points links r={0} />
				<Bars y={(d: TicketBurn) => [d.start, d.end]} radius={.5} />
				<Highlight area />
			  </Svg>
			  <Tooltip.Root let:data>
				<Tooltip.Header>{formatDate(data.date, PeriodType.Day)}</Tooltip.Header>
				<Tooltip.List>
				  <Tooltip.Item label="Shift Start" value={data.start} format="integer" />
				  <Tooltip.Item label="High" value={data.high} format="integer" />
				  <Tooltip.Item label="Low" value={data.low} format="integer" />
				  <Tooltip.Item label="Shift End" value={data.end} format="integer" />
				</Tooltip.List>
			  </Tooltip.Root>
			</Chart>
		  </div>
		  
<pre>Time spent on toil work (by category/label/service?)</pre>
	</div>
</div>