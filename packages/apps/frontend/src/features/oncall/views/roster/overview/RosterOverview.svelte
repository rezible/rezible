<script lang="ts">
	import RiBarChart2Line from "remixicon-svelte/icons/bar-chart-2-line";
	import { getOncallRosterMetricsOptions } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import Header from "$src/components/layout/header/Header.svelte";
	import RosterActivityColumn from "./RosterActivityColumn.svelte";
	import RosterShiftsColumn from "./RosterShiftsColumn.svelte";
	import { useOncallRosterViewController } from "$features/oncall/views/roster";

	const view = useOncallRosterViewController();

	let periodDays = $state(30);
	const metricsQuery = createQuery(() => ({
		...getOncallRosterMetricsOptions({ query: { rosterId: view.rosterId } }),
		enabled: !!view.rosterId,
	}));
	const metrics = $derived(metricsQuery.data?.data);
</script>

<div class="w-full h-full grid grid-cols-4 gap-2">
	<div class="col-span-2 h-full w-full overflow-y-auto pr-1 space-y-2">
		<div class="p-2 flex flex-col gap-2 border border-surface-content/10 rounded p-2">
			<Header title="Key Metrics" subheading="Last 30 days" classes={{ root: "text-lg font-medium" }}>
				{#snippet avatar()}
					<RiBarChart2Line aria-hidden="true" />
				{/snippet}
			</Header>

			{#if metrics}
				<div class="flex gap-2 flex-wrap">
					<!-- <MetricCard
						title="Health Score"
						icon={RiHeartPulseLine}
						metric={metrics.healthScore}
						comparison={{ value: .44 }}
					/>
					<MetricCard title="Incidents" icon={RiFireLine} metric={metrics.incidents} />
					<MetricCard title="Alerts" icon={RiNotification2Line} metric={metrics.alerts} />
					<MetricCard
						title="Alert Actionability"
						icon={RiNotificationBadgeLine}
						metric="{metrics.alertActionability * 100}%"
					/> -->
				</div>
			{/if}
		</div>
	</div>

	<div class="h-full flex flex-col overflow-y-auto">
		<RosterShiftsColumn />
	</div>

	<div class="h-full flex flex-col overflow-y-auto">
		<RosterActivityColumn />
	</div>
</div>
