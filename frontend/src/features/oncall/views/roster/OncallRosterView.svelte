<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallRosterOptions, type OncallRoster } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import OncallRosterDetails from "./OncallRosterDetails.svelte";

	type Props = { rosterId: string };
	const { rosterId }: Props = $props();

	const query = createQuery(() => getOncallRosterOptions({ path: { id: rosterId } }));

	const rosterName = $derived(query.data?.data.attributes.name);

	setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Rosters", href: "/oncall/rosters" },
		{
			label: rosterName,
			href: `/oncall/rosters/${rosterId}`,
			avatar: { kind: "roster", id: rosterId },
		},
	]);
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(roster: OncallRoster)}
		<OncallRosterDetails {roster} />
	{/snippet}
</LoadingQueryWrapper>
