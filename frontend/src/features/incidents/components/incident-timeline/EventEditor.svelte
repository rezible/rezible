<script lang="ts">
	import { Button, ExpansionPanel, Icon, ListItem, TextField, autoHeight } from 'svelte-ux';
	import { mdiFileQuestion, mdiGithub, mdiSlack } from '@mdi/js';
	import type { EventData, IncidentEvent } from './events';
	import Slack from './DataSources/Slack.svelte';
	import Github from './DataSources/Github.svelte';
	import type { Incident } from '$lib/api';

	type Props = { 
		incidentId: string;
		event: IncidentEvent;
		changed: boolean;
	};
	let { incidentId, event, changed = $bindable(false) }: Props = $props();

	const date = $derived(event.start);
	const timeFmt = $derived(`${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`);

	const onChange = () => {
		// TODO
		changed = true;
	};

	const onDateTimeChange = () => {
		const [hoursStr, minutesStr] = timeFmt.split(':');
		const paddedHours = hoursStr.padStart(2, '0');
		const paddedMinutes = minutesStr.padStart(2, '0');
		// const hours = Number.parseInt(hoursStr);
		// const minutes = Number.parseInt(minutesStr);
		const newDate = date.toDateString();
		console.log(`new date: ${newDate} ${paddedHours}:${paddedMinutes}`);
	};

	const dataSources = [
		{ id: 'slack', name: 'Slack', icon: mdiSlack, component: Slack },
		{ id: 'github', name: 'Github', icon: mdiGithub, component: Github }
	];

	let sourceAmounts: { [name: string]: number } = $state({});
	$effect(() => {
		if (event.data) {
			sourceAmounts = {};
			event.data.forEach((v) => {
				sourceAmounts[v.source] = sourceAmounts[v.source] + 1 || 1;
			});
		}
	});

	const toggleDataItem = (e: CustomEvent) => {
		const data = e.detail as EventData;
		const curIndex = event.data?.findIndex((v) => v.id === data.id && v.source === data.source) || -1;
		if (curIndex >= 0) {
			// event.data.splice(curIndex, 1);
		} else {
			// event.data.push(data);
		}
		event.data = event.data;
		onChange();
	};

	// TODO
</script>

<div class="min-h-0 flex-1 overflow-auto flex">
	<div class="w-96">
		{#each dataSources as src, i}
			<ExpansionPanel popout={false}>
				<ListItem
					slot="trigger"
					title={src.name}
					subheading="foo"
					icon={src.icon}
					avatar={{ class: 'bg-surface-content/50 text-surface-100/90' }}
					class="flex-1"
					noShadow
				/>
				<div>
					<span>TODO: render data component</span>
				</div>
			</ExpansionPanel>
		{/each}
	</div>

	<div class="flex flex-col gap-2 p-3 w-96">
		<TextField label="Title" value="" on:change={onChange} />
		<TextField
			label="Time"
			mask="dd/mm/yyyy hh:mm"
			replace="dmyh"
			value={timeFmt}
			on:change={onDateTimeChange}
		/>
		<TextField
			label="Description"
			value=""
			multiline
			actions={(node) => {
				// @ts-expect-error
				return [autoHeight(node)];
			}}
			on:change={onChange}
		/>

		<div class:hidden={event.data?.length === 0} class="flex flex-col">
			<span class="text-sm">Attached Data:</span>
			{#each Object.entries(sourceAmounts) as [id, amt]}
				{#if amt > 0}
					{@const src = dataSources.find((s) => s.id === id)}
					<span class="flex items-center gap-2 text-sm">
						<Icon data={src?.icon || mdiFileQuestion} />
						{amt}
						{id === 'slack' ? 'message' : 'item'}
					</span>
				{/if}
			{/each}
		</div>
	</div>
</div>
