<script lang="ts">
	import RiFilterLine from "remixicon-svelte/icons/filter-line";
	import RiFireLine from "remixicon-svelte/icons/fire-line";
	import RiTodoLine from "remixicon-svelte/icons/todo-line";
	import RiBookOpenLine from "remixicon-svelte/icons/book-open-line";
	import RiShakeHandsLine from "remixicon-svelte/icons/shake-hands-line";
	import RiErrorWarningLine from "remixicon-svelte/icons/error-warning-line";
	import RiArrowRightLine from "remixicon-svelte/icons/arrow-right-line";
	import { Button } from "$components/ui/button";
	import Header from "$src/components/layout/header/Header.svelte";
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

	const recentActivity = $state<ActivityItem[]>([]);

	const getActivityIcon = (type: string) => {
		switch (type) {
			case "incident":
				return RiFireLine;
			case "handover":
				return RiShakeHandsLine;
			case "playbook":
				return RiBookOpenLine;
			case "backlog":
				return RiTodoLine;
			default:
				return RiErrorWarningLine;
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
	};

	const formatDateRelative = (date: Date): string => {
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

		if (diffDays === 0) return "Today";
		if (diffDays === 1) return "Yesterday";
		if (diffDays < 7) return `${diffDays} days ago`;
		return formatRelative(date, Date.now());
	};
</script>

<div class="flex flex-col h-full border border-foreground/10 rounded">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Recent Activity" classes={{ root: "", title: "text-xl" }}>
			{#snippet actions()}
				<Button href={`/rosters/${rosterId}/activity`}>
					View All
					<RiArrowRightLine aria-hidden="true" />
				</Button>
			{/snippet}
		</Header>
	</div>

	<div class="flex-1 flex flex-col px-0 overflow-y-auto">
		{#each recentActivity as activity}
			{@const ActivityIcon = getActivityIcon(activity.type)}
			<div class="p-4 flex items-start gap-3 border-b first:border-t">
				<div class="mt-1">
					<ActivityIcon aria-hidden="true" />
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
