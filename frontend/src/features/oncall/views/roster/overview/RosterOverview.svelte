<script lang="ts">
	import {
		mdiBellAlert,
		mdiBellBadge,
		mdiChartBar,
		mdiFire,
		mdiClipboardList,
		mdiBookOpenVariant,
		mdiHandshake,
		mdiAlertCircle,
		mdiArrowRight,
		mdiHeartPulse,
	} from "@mdi/js";
	import { Button, Header, Icon, Badge, Progress } from "svelte-ux";
	import { formatDistanceToNow, formatRelative } from "date-fns";
	import MetricCard from "$components/viz/MetricCard.svelte";
	import { rosterViewCtx } from "../context.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import {
		getOncallRosterMetricsOptions,
		getUserOncallInformationOptions,
		type OncallShift,
		type User,
	} from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { cls } from "@layerstack/tailwind";
	import { parseAbsoluteToLocal } from "@internationalized/date";

	type ActivityItem = {
		id: string;
		type: "incident" | "handover" | "playbook" | "backlog";
		title: string;
		timestamp: Date;
		user?: User;
	};

	const mockRecentActivity: ActivityItem[] = [
		{
			id: "incident-1",
			type: "incident",
			title: "Database Outage",
			timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24),
			user: {
				id: "user-2",
				attributes: {
					name: "User Name",
					email: "user@example.com",
				},
			},
		},
		{
			id: "handover-1",
			type: "handover",
			title: "Weekly Handover Completed",
			timestamp: new Date(Date.now() - 1000 * 60 * 60 * 48),
			user: {
				id: "user-1",
				attributes: {
					name: "User Name",
					email: "user@example.com",
				},
			},
		},
		{
			id: "playbook-1",
			type: "playbook",
			title: "API Outage Playbook Updated",
			timestamp: new Date(Date.now() - 1000 * 60 * 60 * 72),
			user: {
				id: "user-3",
				attributes: {
					name: "User Name",
					email: "user@example.com",
				},
			},
		},
		{
			id: "backlog-1",
			type: "backlog",
			title: "Improve Alert Filtering",
			timestamp: new Date(Date.now() - 1000 * 60 * 60 * 96),
			user: {
				id: "user-3",
				attributes: {
					name: "User Name",
					email: "user@example.com",
				},
			},
		},
	];

	const viewCtx = rosterViewCtx.get();
	const rosterId = $derived(viewCtx.rosterId);

	let periodDays = $state(30);
	const metricsQuery = createQuery(() => getOncallRosterMetricsOptions({ query: { rosterId } }));
	const metrics = $derived(metricsQuery.data?.data);

	// TODO: use correct query
	const shiftsQuery = createQuery(() => getUserOncallInformationOptions({ query: {} }));
	const shifts = $derived(shiftsQuery.data?.data);
	const prevShift = $derived(shifts?.pastShifts.at(0));
	const activeShift = $derived(shifts?.activeShifts.at(0));
	const nextShift = $derived(shifts?.upcomingShifts.at(0));

	const recentActivity = $state<ActivityItem[]>(mockRecentActivity);

	const getActivityIcon = (type: string) => {
		switch (type) {
			case "incident":
				return mdiFire;
			case "handover":
				return mdiHandshake;
			case "playbook":
				return mdiBookOpenVariant;
			case "backlog":
				return mdiClipboardList;
			default:
				return mdiAlertCircle;
		}
	};

	const getActivityColor = (type: string) => {
		switch (type) {
			case "incident":
				return "text-red-500";
			case "handover":
				return "text-blue-500";
			case "playbook":
				return "text-purple-500";
			case "backlog":
				return "text-gray-500";
			default:
				return "text-gray-500";
		}
	}

	const formatDateRelative = (date: Date): string => {
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

		if (diffDays === 0) return "Today";
		if (diffDays === 1) return "Yesterday";
		if (diffDays < 7) return `${diffDays} days ago`;
		return formatRelative(date, Date.now());
	}
</script>

<div class="flex flex-col gap-4">
	<div class="p-2 flex flex-col gap-2 border border-surface-content/10 rounded p-2">
		<Header title="Key Metrics" subheading="Last 30 days" classes={{ root: "text-lg font-medium" }}>
			<div slot="avatar">
				<Icon data={mdiChartBar} size={38} class="text-primary-300" />
			</div>
		</Header>

		{#if metrics}
		<div class="flex gap-2 flex-wrap">
			<MetricCard
				title="Health Score"
				icon={mdiHeartPulse}
				metric={metrics.healthScore}
				comparison={{ value: .44 }}
			/>
			<MetricCard title="Incidents" icon={mdiFire} metric={metrics.incidents} />
			<MetricCard title="Alerts" icon={mdiBellAlert} metric={metrics.alerts} />
			<MetricCard
				title="Alert Actionability"
				icon={mdiBellBadge}
				metric="{metrics.alertActionability * 100}%"
			/>
		</div>
		{/if}
	</div>

	<div class="flex flex-col gap-2 border border-surface-content/10 rounded p-2">
		<Header title="Oncall Shifts" classes={{ root: "text-lg font-medium" }} />

		{#snippet shiftCard(shift: OncallShift, status: "previous" | "active" | "next")}
			{@const user = shift.attributes.user}
			{@const isActive = status === "active"}
			<div
				class={cls(
					"border rounded-md p-4 flex flex-col gap-2 min-w-72",
					isActive
						? "border-success-900 bg-success-900/10"
						: "border-neutral-content/10 bg-neutral-900/30"
				)}
			>
				<div class="flex items-center gap-2">
					<Avatar id={user.id} kind="user" size={30} />
					<span class="text-lg">{user.attributes.name}</span>
				</div>

				<div class="flex flex-col">
					{#if status === "previous"}
						<span
							>Ended {formatDistanceToNow(
								parseAbsoluteToLocal(shift.attributes.endAt).toDate()
							)} ago</span
						>
					{:else if status === "active"}
						<span>Active Now</span>
					{:else}
						<span
							>Starts in {formatDistanceToNow(
								parseAbsoluteToLocal(shift.attributes.startAt).toDate()
							)}</span
						>
					{/if}
				</div>

				<div class="h-8 flex items-center self-end">
					<Button
						variant="fill"
						color={isActive ? "success" : "neutral"}
						href={`/oncall/shifts/${shift.id}`}
					>
						View
					</Button>
				</div>
			</div>
		{/snippet}
		<div class="flex gap-4 items-center">
			{#if prevShift}{@render shiftCard(prevShift, "previous")}{/if}
			{#if activeShift}{@render shiftCard(activeShift, "active")}{/if}
			{#if nextShift}{@render shiftCard(nextShift, "next")}{/if}
		</div>
	</div>

	<div class="flex flex-col gap-2 border border-surface-content/10 rounded p-2">
		<Header title="Recent Activity" classes={{ root: "text-lg font-medium" }}>
			<svelte:fragment slot="actions"></svelte:fragment>
		</Header>

		<div class="flex flex-col border rounded-md divide-y">
			{#each recentActivity as activity}
				<div class="p-4 flex items-start gap-3">
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

						<div class="mt-2">
							<Button
								variant="fill"
								color="neutral"
								size="sm"
								href={`/oncall/rosters/${rosterId}/backlog/${activity.id}`}
							>
								View
							</Button>
						</div>
					</div>
				</div>
			{/each}
		</div>

		<Button variant="fill-light" href={`/oncall/rosters/${rosterId}/activity`}>
			View All Activity
			<Icon data={mdiArrowRight} class="ml-1 h-4 w-4" />
		</Button>
	</div>
</div>
