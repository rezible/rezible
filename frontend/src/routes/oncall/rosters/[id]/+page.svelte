<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { page } from "$app/state";
	import { getOncallRosterOptions, type OncallRoster } from "$lib/api";
	import PageContainer, {
		type Breadcrumb,
	} from "$components/page-container/PageContainer.svelte";
	import RosterView from "$features/oncall/views/roster/RosterView.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";

	const query = createQuery(() =>
		getOncallRosterOptions({ path: { id: page.params.id } })
	);
	const roster = $derived(query.data?.data);

	const breadcrumbs = $derived<Breadcrumb[]>([
		{ label: "Oncall", href: "/oncall" },
		{ label: "Rosters", href: "/oncall/rosters" },
		{
			label: roster?.attributes.name ?? "",
			avatar: roster ? { kind: "roster", id: roster.id } : undefined,
		},
	]);
</script>

<PageContainer {breadcrumbs}>
	<LoadingQueryWrapper {query}>
		{#snippet view(roster: OncallRoster)}
			<RosterView {roster} />
		{/snippet}
	</LoadingQueryWrapper>
</PageContainer>
