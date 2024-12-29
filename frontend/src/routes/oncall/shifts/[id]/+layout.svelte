<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
    import { page } from "$app/state";
    import { getOncallShiftOptions, type OncallShift } from "$lib/api";
    import PageContainer, { type PageTabsProps, type Breadcrumb } from "$components/page-container/PageContainer.svelte";

	const { children } = $props();

	const shiftId = $derived(page.params.id);
	const query = createQuery(() => getOncallShiftOptions({path: {id: shiftId}}));
	const shift = $derived(query.data?.data);

	const formatShiftDates = (shift: OncallShift) => {
		const start = new Date(shift.attributes.start_at);
		const end = new Date(shift.attributes.end_at);
		const rosterName = shift.attributes.roster.attributes.name;
		return `${rosterName} - ${start.toDateString()} to ${end.toDateString()}`
	}

	const isHandover = $derived(page.route.id === "/oncall/shifts/[id]/handover");

	const baseCrumbs: Breadcrumb[] = [
		{label: "Oncall", href: "/oncall"},
		{label: "Shifts", href: "/oncall/shifts"},
	];
	const shiftCrumb = $derived<Breadcrumb>(shift ? {label: formatShiftDates(shift), href: "/oncall/shifts/" + shiftId} : {label: ""});
	const handoverCrumb = $derived<Breadcrumb>({label: "Handover", href: `/oncall/shifts/${shiftId}/handover`});
	const shiftCrumbs: Breadcrumb[] = $derived(isHandover ? [shiftCrumb, handoverCrumb] : [shiftCrumb]);
	
	const breadcrumbs = $derived<Breadcrumb[]>([...baseCrumbs, ...shiftCrumbs]);
</script>

<PageContainer {breadcrumbs}>
	{@render children()}
</PageContainer>