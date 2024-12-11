<script lang="ts">
	import {
		mdiClipboard,
		mdiFileCheck,
		mdiOpenInNew,
		mdiProgressPencil,
		mdiVideoAccount
	} from '@mdi/js';
	import { Button, Header, Icon } from 'svelte-ux';
    import { createQuery } from '@tanstack/svelte-query';
	import { getRetrospectiveForIncidentOptions, type Incident } from '$lib/api';

	interface Props { incident: Incident };
	let { incident }: Props = $props();

	const retrospectiveQuery = createQuery(() => ({
		...getRetrospectiveForIncidentOptions({path: {id: incident.id}}),
	}));
	const retrospective = $derived(retrospectiveQuery.data?.data);

	const status = $derived.by(() => {
		switch (retrospective?.attributes.status) {
			case 'open':
				return { label: 'In Progress', icon: mdiProgressPencil, color: 'text-info-300' };
			case 'in_review':
				return { label: 'Awaiting Review', icon: mdiClipboard, color: 'text-accent-200' };
			case 'meeting_scheduled':
				return {
					label: 'Incident Review Meeting Scheduled',
					icon: mdiVideoAccount,
					color: 'text-success-300'
				};
			case 'completed':
				return { label: 'Completed', icon: mdiFileCheck, color: 'text-success-600' };
		}
		return { label: "Loading", icon: mdiClipboard, color: "text-neutral" };
	});
</script>

<a
	class="flex flex-col gap-2 border rounded-lg p-2 bg-surface-200 hover:bg-surface-300 group"
	href="/incidents/{incident.attributes.slug}/retrospective"
>
	<div class="">
		<Header title="Incident Retrospective" classes={{ title: 'text-lg text-neutral-50' }}>
			<div slot="actions" class="">
				<Button
					color="primary"
					variant="text"
					classes={{ root: 'text-primary-content' }}
					href="/incidents/{incident.attributes.slug}/retrospective"
				>
					Open
					<Icon data={mdiOpenInNew} />
				</Button>
			</div>
		</Header>
		<span class="flex items-center gap-2 text-md {status.color}">
			<Icon data={status.icon} />
			{status.label}
		</span>
	</div>

	<div
		class="border border-surface-content/10 bg-surface-100 group-hover:bg-surface-200 rounded-lg shadow-lg p-2"
	>
		<Header title="Summary" classes={{ title: 'text-neutral-50' }} />
		<p class="text-sm">{incident.attributes.summary}</p>
	</div>
</a>