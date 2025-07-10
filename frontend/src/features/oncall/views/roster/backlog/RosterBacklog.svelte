<script lang="ts">
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import { getLocalTimeZone, now } from "@internationalized/date";

	import { rosterViewCtx } from "../viewState.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import Header from "$src/components/header/Header.svelte";
	
	type BacklogItem = {
		id: string;
		attributes: BacklogItemAttributes;
	};

	type BacklogItemAttributes = {
		title: string;
		priority: number;
		createdAt: Date;
	};

	const viewCtx = rosterViewCtx.get();
	const rosterId = $derived(viewCtx.rosterId);

	const mockBacklogItems: BacklogItem[] = [
		// TODO: fill this with some items
	];

	const listRosterBacklogItemsOptions = (options: {query: {rosterId: string}}) => {
		return queryOptions({
			queryFn: async ({ queryKey, signal }) => {
				return {data: mockBacklogItems};
			},
			queryKey: ["getUserOncallInformationOptions", options.query.rosterId]
		});
	};
	const backlogQuery = createQuery(() => listRosterBacklogItemsOptions({query: {rosterId}}));

	const today = now(getLocalTimeZone());

	type TicketBurn = { date: Date, start: number, high: number, low: number, end: number };
	const mockBurndownDay = (weeks: number, start: number, high: number, low: number, end: number): TicketBurn => {
		return {start, high, low, end, date: today.subtract({weeks}).toDate()}
	}
	const burndownData: TicketBurn[] = [
		mockBurndownDay(4, 3, 8, 3, 4),
		mockBurndownDay(3, 4, 6, 3, 3),
		mockBurndownDay(2, 3, 11, 3, 6),
		mockBurndownDay(1, 6, 14, 6, 10),
		mockBurndownDay(0, 10, 11, 6, 6),
	]
</script>

<div class="grid grid-cols-2 gap-2 h-full">
	<div class="flex flex-col">
		<Header title="Tickets" classes={{root: "text-lg font-medium"}} />

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
		<Header title="Ticket Burndown" classes={{root: "text-lg font-medium"}} />

		<div class="h-[300px] p-4 border rounded">
			<!--Chart
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
			</Chart-->
		</div>
		
		<!-- <pre>Time spent on toil work (by category/label/service?)</pre> -->
	</div>
</div>