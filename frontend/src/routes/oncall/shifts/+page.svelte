<script>
	import PageContainer from "$components/page-container/PageContainer.svelte";
	import ShiftsList from "$features/oncall/views/shifts-list/ShiftsList.svelte";
</script>

<PageContainer
	breadcrumbs={[{ label: "Oncall", href: "/oncall" }, { label: "Shifts" }]}
>
	<ShiftsList />
</PageContainer>

<!-- upcoming -->
<!--script lang="ts">
	import type { OncallShift } from '$lib/api';
	import { formatDistanceToNow } from 'date-fns';
	import { Card, Header } from 'svelte-ux';
	// import { oncall } from '$lib/api';
	// import { createQuery } from '@tanstack/svelte-query';
	// const userOncallQuery = createQuery(oncall.getUserDetails());
    
	const shifts: OncallShift[] = [];
</script>

<div class="flex flex-row gap-2 overflow-y-auto">
	{#each shifts as shift}
	{@const attr = shift.attributes}
	{@const isPrimary = false}
		<Card class="w-full" classes={{ headerContainer: 'py-2' }}>
			<div slot="header">
				<Header title={attr.roster.attributes.name}>
					<div slot="title">
						<span class="text-lg">roster</span>
					</div>
					<div slot="subheading">
						<span class="text-sm text-surface-content/75">
							{isPrimary ? 'Primary' : 'Secondary'}
						</span>
					</div>
					<div slot="actions"></div>
				</Header>
			</div>

			<div slot="contents" class="pb-2">
				<span>Starts in {formatDistanceToNow(new Date(attr.start))}</span>
			</div>
		</Card>
	{/each}
</div-->
