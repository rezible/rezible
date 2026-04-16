<script lang="ts">
	import type { Incident, IncidentAttributes } from "$lib/api";
	import { mdiAccountGroup, mdiCalendarStart, mdiChevronRight, mdiLock, mdiTag } from "@mdi/js";
	import { Badge } from "$components/ui/badge";
	import Icon from "$components/icon/Icon.svelte";

	type Props = {
		incident: Incident;
	};
	const { incident }: Props = $props();

	type IncidentStatus = IncidentAttributes["currentStatus"];

	const statusLabels: Record<IncidentStatus, string> = {
		started: "Started",
		mitigated: "Mitigated",
		resolved: "Resolved",
		closed: "Closed",
	};

	const statusClasses: Record<IncidentStatus, string> = {
		started: "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300",
		mitigated: "border-sky-500/40 bg-sky-500/10 text-sky-700 dark:text-sky-300",
		resolved: "border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
		closed: "border-muted-foreground/30 bg-muted text-muted-foreground",
	};

	const attrs = $derived(incident.attributes);
	const status = $derived(normalizeStatus(attrs.currentStatus));
	const openedAtLabel = $derived(formatDate(attrs.openedAt));
	const severityName = $derived(attrs.severity?.attributes.name);
	const typeName = $derived(attrs.type?.attributes.name);
	const activeTeams = $derived(attrs.teams?.filter((assignment) => assignment.active) ?? []);
	const visibleTags = $derived(attrs.tags?.slice(0, 3) ?? []);
	const href = $derived(`/incidents/${attrs.slug || incident.id}`);

	function formatDate(value: string | undefined) {
		if (!value) return "No open date";
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return "No open date";
		return new Intl.DateTimeFormat(undefined, {
			month: "short",
			day: "numeric",
			year: "numeric",
		}).format(date);
	}

	function normalizeStatus(value: string | undefined): IncidentStatus {
		return value && value in statusLabels ? (value as IncidentStatus) : "started";
	}
</script>

<a
	{href}
	class="group grid min-h-28 grid-cols-[minmax(0,1fr)_auto] gap-3 rounded-lg border bg-card px-4 py-3 text-card-foreground shadow-sm transition hover:border-primary/40 hover:bg-muted/30 hover:shadow-md"
>
	<div class="flex min-w-0 flex-col gap-3">
		<div class="flex min-w-0 flex-col gap-2">
			<div class="flex flex-wrap items-center gap-2">
				<Badge variant="outline" class={statusClasses[status]}>
					{statusLabels[status]}
				</Badge>
				{#if severityName}
					<Badge variant="secondary">{severityName}</Badge>
				{/if}
				{#if typeName}
					<Badge variant="outline">{typeName}</Badge>
				{/if}
				{#if attrs.private}
					<Badge variant="outline" class="gap-1">
						<Icon data={mdiLock} size={12} />
						Private
					</Badge>
				{/if}
			</div>

			<div class="min-w-0">
				<div class="truncate text-base font-semibold leading-tight">{attrs.title}</div>
				{#if attrs.summary}
					<p class="mt-1 line-clamp-2 text-sm text-muted-foreground">{attrs.summary}</p>
				{/if}
			</div>
		</div>

		<div class="flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground">
			<span class="inline-flex items-center gap-1">
				<Icon data={mdiCalendarStart} size={14} />
				Opened {openedAtLabel}
			</span>

			{#if activeTeams.length > 0}
				<span class="inline-flex items-center gap-1">
					<Icon data={mdiAccountGroup} size={14} />
					{activeTeams.map((assignment) => assignment.team.attributes.name).join(", ")}
				</span>
			{/if}

			{#if visibleTags.length > 0}
				<span class="inline-flex items-center gap-1">
					<Icon data={mdiTag} size={14} />
					{visibleTags.map((tag) => tag.attributes.value).join(", ")}
					{#if (attrs.tags?.length ?? 0) > visibleTags.length}
						+{(attrs.tags?.length ?? 0) - visibleTags.length}
					{/if}
				</span>
			{/if}
		</div>
	</div>

	<div
		class="flex items-center text-muted-foreground transition group-hover:translate-x-0.5 group-hover:text-foreground"
	>
		<Icon data={mdiChevronRight} size={22} />
	</div>
</a>
