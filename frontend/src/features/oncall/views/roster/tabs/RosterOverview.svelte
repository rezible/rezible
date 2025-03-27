<script lang="ts">
	import { 
		mdiBellAlert, 
		mdiBellBadge, 
		mdiChartBar, 
		mdiFire, 
		mdiPacMan, 
		mdiAccountGroup, 
		mdiClockOutline, 
		mdiClipboardList, 
		mdiBookOpenVariant, 
		mdiHandshake, 
		mdiAlertCircle, 
		mdiCheckCircle, 
		mdiArrowRight, 
		mdiCalendarClock
	} from "@mdi/js";
	import MetricCard from "$components/viz/MetricCard.svelte";
	import { Button, Header, Icon, Avatar, Badge, Progress } from "svelte-ux";
	import TimePeriodSelect from "$components/time-period-select/TimePeriodSelect.svelte";
	import { rosterIdCtx } from "../context";
	import type { HealthIndicator, ActivityItem, OncallShiftDisplay } from "../types";
	import { formatDate, formatTime } from "$lib/utils/date";

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

	// Mock data for oncall shifts
	const now = new Date();
	const yesterday = new Date(now);
	yesterday.setDate(yesterday.getDate() - 1);
	const tomorrow = new Date(now);
	tomorrow.setDate(tomorrow.getDate() + 1);
	const dayAfterTomorrow = new Date(now);
	dayAfterTomorrow.setDate(dayAfterTomorrow.getDate() + 2);

	const shifts = $state<OncallShiftDisplay[]>([
		{
			status: 'previous',
			startTime: yesterday,
			endTime: now,
			shift: {
				id: "shift-1",
				attributes: {
					startAt: yesterday.toISOString(),
					endAt: now.toISOString(),
					role: "Primary",
					user: {
						id: "user-1",
						attributes: {
							name: "Alex Johnson",
							email: "alex@example.com"
						}
					},
					roster: {
						id: rosterId,
						attributes: {
							name: "Platform Team",
							slug: "platform",
							schedules: [],
							handoverTemplateId: "template-1"
						}
					},
					covers: []
				}
			}
		},
		{
			status: 'current',
			startTime: now,
			endTime: tomorrow,
			shift: {
				id: "shift-2",
				attributes: {
					startAt: now.toISOString(),
					endAt: tomorrow.toISOString(),
					role: "Primary",
					user: {
						id: "user-2",
						attributes: {
							name: "Sam Taylor",
							email: "sam@example.com"
						}
					},
					roster: {
						id: rosterId,
						attributes: {
							name: "Platform Team",
							slug: "platform",
							schedules: [],
							handoverTemplateId: "template-1"
						}
					},
					covers: []
				}
			}
		},
		{
			status: 'next',
			startTime: tomorrow,
			endTime: dayAfterTomorrow,
			shift: {
				id: "shift-3",
				attributes: {
					startAt: tomorrow.toISOString(),
					endAt: dayAfterTomorrow.toISOString(),
					role: "Primary",
					user: {
						id: "user-3",
						attributes: {
							name: "Jamie Smith",
							email: "jamie@example.com"
						}
					},
					roster: {
						id: rosterId,
						attributes: {
							name: "Platform Team",
							slug: "platform",
							schedules: [],
							handoverTemplateId: "template-1"
						}
					},
					covers: []
				}
			}
		}
	]);

	// Mock data for health indicators
	const healthIndicators = $state<HealthIndicator[]>([
		{
			name: "Playbook Freshness",
			value: 85,
			target: 90,
			status: 'warning'
		},
		{
			name: "Handover Completion",
			value: 95,
			target: 90,
			status: 'good'
		},
		{
			name: "KTLO Backlog",
			value: 12,
			target: 10,
			status: 'warning'
		},
		{
			name: "Alert Noise Ratio",
			value: 25,
			target: 30,
			status: 'good'
		},
		{
			name: "Team Load",
			value: 70,
			target: 80,
			status: 'good'
		}
	]);

	// Mock data for recent activity
	const recentActivity = $state<ActivityItem[]>([
		{
			id: "incident-1",
			type: "incident",
			title: "Database Outage",
			timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24),
			user: {
				id: "user-2",
				attributes: {
					name: "Sam Taylor",
					email: "sam@example.com"
				}
			}
		},
		{
			id: "handover-1",
			type: "handover",
			title: "Weekly Handover Completed",
			timestamp: new Date(Date.now() - 1000 * 60 * 60 * 48),
			user: {
				id: "user-1",
				attributes: {
					name: "Alex Johnson",
					email: "alex@example.com"
				}
			}
		},
		{
			id: "playbook-1",
			type: "playbook",
			title: "API Outage Playbook Updated",
			timestamp: new Date(Date.now() - 1000 * 60 * 60 * 72),
			user: {
				id: "user-3",
				attributes: {
					name: "Jamie Smith",
					email: "jamie@example.com"
				}
			}
		},
		{
			id: "backlog-1",
			type: "backlog",
			title: "Improve Alert Filtering",
			timestamp: new Date(Date.now() - 1000 * 60 * 60 * 96)
		}
	]);

	function getStatusColor(status: 'good' | 'warning' | 'critical') {
		switch (status) {
			case 'good': return 'bg-green-100 text-green-800';
			case 'warning': return 'bg-yellow-100 text-yellow-800';
			case 'critical': return 'bg-red-100 text-red-800';
			default: return 'bg-gray-100 text-gray-800';
		}
	}

	function getActivityIcon(type: string) {
		switch (type) {
			case 'incident': return mdiFire;
			case 'handover': return mdiHandshake;
			case 'playbook': return mdiBookOpenVariant;
			case 'backlog': return mdiClipboardList;
			default: return mdiAlertCircle;
		}
	}

	function getActivityColor(type: string) {
		switch (type) {
			case 'incident': return 'text-red-500';
			case 'handover': return 'text-blue-500';
			case 'playbook': return 'text-purple-500';
			case 'backlog': return 'text-gray-500';
			default: return 'text-gray-500';
		}
	}

	function formatDateRelative(date: Date): string {
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
		
		if (diffDays === 0) return 'Today';
		if (diffDays === 1) return 'Yesterday';
		if (diffDays < 7) return `${diffDays} days ago`;
		return formatDate(date);
	}
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

		<div class="grid grid-cols-3 gap-4">
			{#each shifts as shiftDisplay}
				<div class="border rounded-md p-4 flex flex-col gap-2 hover:shadow-md transition-shadow">
					<div class="flex justify-between items-center">
						<Badge 
							variant={shiftDisplay.status === 'current' ? 'solid' : 'outline'} 
							color={shiftDisplay.status === 'current' ? 'primary' : 'neutral'}
							class="capitalize"
						>
							{shiftDisplay.status} shift
						</Badge>
						<span class="text-sm text-gray-500">{shiftDisplay.shift.attributes.role}</span>
					</div>
					
					<div class="flex items-center gap-2 mt-2">
						<Avatar 
							name={shiftDisplay.shift.attributes.user.attributes.name} 
							size="md" 
							class="flex-shrink-0"
						/>
						<div class="flex flex-col">
							<span class="font-medium">{shiftDisplay.shift.attributes.user.attributes.name}</span>
							<span class="text-sm text-gray-500">{shiftDisplay.shift.attributes.user.attributes.email}</span>
						</div>
					</div>
					
					<div class="flex items-center gap-1 mt-2 text-sm text-gray-600">
						<Icon data={mdiCalendarClock} class="h-4 w-4" />
						<div class="flex flex-col">
							<div class="flex items-center gap-1">
								<span>From:</span>
								<span class="font-medium">{formatDate(shiftDisplay.startTime)} {formatTime(shiftDisplay.startTime)}</span>
							</div>
							<div class="flex items-center gap-1">
								<span>To:</span>
								<span class="font-medium">{formatDate(shiftDisplay.endTime)} {formatTime(shiftDisplay.endTime)}</span>
							</div>
						</div>
					</div>
					
					{#if shiftDisplay.status === 'current'}
						<Button 
							variant="outline" 
							class="mt-2" 
							href={`/oncall/rosters/${rosterId}/handover`}
						>
							Prepare Handover
						</Button>
					{:else if shiftDisplay.status === 'next'}
						<Button 
							variant="outline" 
							class="mt-2" 
							href={`/oncall/rosters/${rosterId}/schedule`}
						>
							View Schedule
						</Button>
					{:else}
						<Button 
							variant="outline" 
							class="mt-2" 
							href={`/oncall/rosters/${rosterId}/history`}
						>
							View History
						</Button>
					{/if}
				</div>
			{/each}
		</div>
	</div>

	<div class="border p-2 flex flex-col gap-2">
		<Header title="Health Indicators" classes={{root: "text-lg font-medium"}} />

		<div class="grid grid-cols-5 gap-4">
			{#each healthIndicators as indicator}
				<div class="border rounded-md p-4 flex flex-col gap-2">
					<div class="flex justify-between items-center">
						<span class="font-medium">{indicator.name}</span>
						<Badge 
							variant="solid" 
							color={indicator.status === 'good' ? 'success' : indicator.status === 'warning' ? 'warning' : 'error'}
							class="capitalize"
						>
							{indicator.status}
						</Badge>
					</div>
					
					<div class="text-2xl font-bold mt-1">
						{indicator.value}{indicator.name === 'Team Load' || indicator.name === 'Playbook Freshness' || indicator.name === 'Handover Completion' ? '%' : ''}
					</div>
					
					<div class="mt-2">
						<Progress 
							value={indicator.value} 
							max={indicator.name === 'KTLO Backlog' ? indicator.target * 2 : 100} 
							class={indicator.status === 'good' ? 'bg-green-500' : indicator.status === 'warning' ? 'bg-yellow-500' : 'bg-red-500'}
						/>
					</div>
					
					<div class="text-sm text-gray-500 mt-1 flex justify-between">
						<span>Target: {indicator.target}{indicator.name === 'Team Load' || indicator.name === 'Playbook Freshness' || indicator.name === 'Handover Completion' ? '%' : ''}</span>
						<span>{indicator.status === 'good' ? 'On track' : indicator.status === 'warning' ? 'Needs attention' : 'Critical'}</span>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<div class="border p-2 flex flex-col gap-2">
		<Header title="Recent Activity" classes={{root: "text-lg font-medium"}} />

		<div class="flex flex-col border rounded-md divide-y">
			{#each recentActivity as activity}
				<div class="p-4 flex items-start gap-3 hover:bg-gray-50">
					<div class="mt-1">
						<Icon 
							data={getActivityIcon(activity.type)} 
							class={`h-5 w-5 ${getActivityColor(activity.type)}`} 
						/>
					</div>
					
					<div class="flex-1">
						<div class="flex flex-col">
							<div class="font-medium">{activity.title}</div>
							<div class="text-sm text-gray-500 flex items-center gap-1">
								<span>{formatDateRelative(activity.timestamp)}</span>
								{#if activity.user}
									<span>•</span>
									<span>{activity.user.attributes.name}</span>
								{/if}
							</div>
						</div>
						
						{#if activity.type === 'incident'}
							<div class="mt-2">
								<Button 
									variant="outline" 
									size="sm" 
									href={`/incidents/${activity.id}`}
								>
									View Incident
								</Button>
							</div>
						{:else if activity.type === 'handover'}
							<div class="mt-2">
								<Button 
									variant="outline" 
									size="sm" 
									href={`/oncall/rosters/${rosterId}/handovers/${activity.id}`}
								>
									View Handover
								</Button>
							</div>
						{:else if activity.type === 'playbook'}
							<div class="mt-2">
								<Button 
									variant="outline" 
									size="sm" 
									href={`/playbooks/${activity.id}`}
								>
									View Playbook
								</Button>
							</div>
						{:else if activity.type === 'backlog'}
							<div class="mt-2">
								<Button 
									variant="outline" 
									size="sm" 
									href={`/oncall/rosters/${rosterId}/backlog/${activity.id}`}
								>
									View Item
								</Button>
							</div>
						{/if}
					</div>
				</div>
			{/each}
			
			<div class="p-4">
				<Button 
					variant="outline" 
					class="w-full" 
					href={`/oncall/rosters/${rosterId}/activity`}
				>
					View All Activity
					<Icon data={mdiArrowRight} class="ml-1 h-4 w-4" />
				</Button>
			</div>
		</div>
	</div>

</div>
