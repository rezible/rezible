<script lang="ts">
	import { mdiBellAlert, mdiBellBadge, mdiChartBar, mdiFire, mdiPacMan } from "@mdi/js";
	import MetricCard from "$components/viz/MetricCard.svelte";
	import { Button, Header, Icon } from "svelte-ux";
	import TimePeriodSelect from "$components/time-period-select/TimePeriodSelect.svelte";
	import { rosterIdCtx } from "../context";

	const rosterId = rosterIdCtx.get();

	type RosterMetrics = {
		incidents: number;
		alerts: number;
		alertActionability: number;
		outOfHoursAlerts: number;
		oncallBurden: number;
		backlogBurnRate: number;
	};

	let periodDays = $state(30);
	const metrics = $derived<RosterMetrics>({
		incidents: 2,
		alerts: 3,
		alertActionability: 0.4,
		outOfHoursAlerts: 8,
		oncallBurden: 64,
		backlogBurnRate: 1.1,
	});
</script>

<div class="flex flex-col gap-4">
	
	<div class="border p-2 flex flex-col gap-2">
		<Header title="Key Metrics" classes={{root: "text-lg font-medium w-fit"}}>
			<div slot="avatar">
				<Icon data={mdiChartBar} class="text-primary-300" />
			</div>
			<div class="" slot="actions">
				<TimePeriodSelect bind:selected={periodDays} />
			</div>
		</Header>

		<div class="flex gap-2 flex-wrap">
			<MetricCard title="Incidents" icon={mdiFire} metric={metrics.incidents} />
			<MetricCard title="Alerts" icon={mdiBellAlert} metric={metrics.alerts} />
			<MetricCard title="Alert Actionability %" icon={mdiBellBadge} metric={metrics.alertActionability} />
			<MetricCard title="Oncall Burden" icon={mdiPacMan} metric={metrics.oncallBurden} />
		</div>

		<!-- <Button variant="default" href="/oncall/rosters/{rosterId}/analysis">View Analytics</Button> -->
	</div>

	<div class="border p-2 flex flex-col gap-2">
		<Header title="Oncall Shifts" classes={{root: "text-lg font-medium"}} />

		<div class="grid grid-cols-3">
			<div class="border">
				previous shift
			</div>
			<div class="border">
				current shift
			</div>
			<div class="border">
				next shift
			</div>
		</div>
	</div>

	<div class="border p-2 flex flex-col gap-2">
		<Header title="Health Indicators" classes={{root: "text-lg font-medium"}} />

		<div class="grid grid-cols-5">
			<div class="border">
				Playbook freshness
			</div>
			<div class="border">
				Handover completion rate
			</div>
			<div class="border">
				KTLO backlog size
			</div>
			<div class="border">
				Alert noise ratio
			</div>
			<div class="border">
				Team load
			</div>
		</div>
	</div>

	<div class="border p-2 flex flex-col gap-2">
		<Header title="Recent Activity" classes={{root: "text-lg font-medium"}} />

		<div class="flex flex-col border">
			<!-- 
				Major incidents (last 7 days)
				Completed handovers
				Updated playbooks
				New KTLO items 
			-->
		</div>
	</div>

</div>