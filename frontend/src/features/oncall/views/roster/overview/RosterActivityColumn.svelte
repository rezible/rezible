<script lang="ts">
	import { 
		mdiFilter,
		mdiFire,
		mdiClipboardList,
		mdiBookOpenVariant,
		mdiHandshake,
		mdiAlertCircle,
		mdiArrowRight,
	} from "@mdi/js";
	import { Button } from "$components/ui/button";
	import Header from "$components/header/Header.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import type { User } from "$lib/api";
	import { formatRelative } from "date-fns";
	import { useOncallRosterViewController } from "$features/oncall/views/roster";
	
	const view = useOncallRosterViewController();
	const rosterId = $derived(view.rosterId);

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
					isOrgAdmin: false,
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
					isOrgAdmin: false,
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
					isOrgAdmin: false,
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
					isOrgAdmin: false,
				},
			},
		},
	];

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

<div class="flex flex-col h-full border border-surface-content/10 rounded">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Recent Activity" classes={{root: "", title: "text-xl"}}>
			{#snippet actions()}
				<Button href={`/rosters/${rosterId}/activity`}>
					View All
					<Icon data={mdiArrowRight} classes={{root: "ml-1 h-4 w-4"}} />
				</Button>
			{/snippet}
		</Header>
	</div>

	<div class="flex-1 flex flex-col px-0 overflow-y-auto">
		{#each recentActivity as activity}
			<div class="p-4 flex items-start gap-3 border-b first:border-t">
				<div class="mt-1">
					<Icon
						data={getActivityIcon(activity.type)}
						classes={{root: `h-5 w-5 ${getActivityColor(activity.type)}`}}
					/>
				</div>

				<div class="flex-1 flex justify-between">
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

					<div class="mt-2 flex self-center">
						<Button
							color="neutral"
							size="sm"
							href={`/rosters/${rosterId}/backlog/${activity.id}`}
						>
							View
						</Button>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
