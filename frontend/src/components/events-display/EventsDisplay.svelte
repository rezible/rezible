<script lang="ts" module>
	export type DisplayEvent = {
		title: string;
		start: Date;
		end: Date;
		color?: string;
	};
</script>

<script lang="ts">
	type Props = {
		events: Array<DisplayEvent>;
		startDate?: Date;
	};

	const { events, startDate = new Date() }: Props = $props();

	const weekDays = $derived.by(() => {
		const currentDate = new Date(startDate);
		currentDate.setDate(currentDate.getDate() - currentDate.getDay());

		const days = [];
		for (let i = 0; i < 7; i++) {
			const date = new Date(currentDate);
			date.setDate(currentDate.getDate() + i);
			const dayEvents = events.filter((event) =>
				isSameDay(event.start, date)
			);
			days.push({ date, dayEvents });
		}
		return days;
	});

	const isSameDay = (date1: Date, date2: Date) => {
		return (
			date1.getFullYear() === date2.getFullYear() &&
			date1.getMonth() === date2.getMonth() &&
			date1.getDate() === date2.getDate()
		);
	};

	const calculateEventDuration = (event: { start: Date; end: Date }) => {
		return (event.end.getTime() - event.start.getTime()) / (1000 * 60 * 60);
	};

	const getDayName = (date: Date) => {
		return date.toLocaleDateString("en-US", { weekday: "short" });
	};
</script>

<div class="grid grid-cols-7 gap-1 p-2 bg-surface-100">
	{#each weekDays as day}
		<div
			class="surface-100 border rounded-container-token p-2 min-h-[120px]"
		>
			<div class="text-xs font-medium text-surface-600 mb-2">
				{getDayName(day.date)}
				{day.date.getDate()}
			</div>
			<div class="space-y-1">
				{#each day.dayEvents as event}
					<div
						style={`height: ${Math.max(calculateEventDuration(event) * 15, 30)}px`}
						class="border rounded-lg bg-surface-200"
					></div>
				{/each}
			</div>
		</div>
	{/each}
</div>
