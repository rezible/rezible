<script lang="ts">
	import type { Incident } from '$lib/api';
	import { Checkbox, Icon, ListItem, Tabs } from 'svelte-ux';
	import { mdiAccountSearch } from '@mdi/js';

	interface Props { 
		incident: Incident;
		data: any[];
	};
	let { incident, data }: Props = $props();

	let checked: { [id: string]: boolean } = $state({});
	$effect(() => {
		if (data) {
			checked = {};
			data.filter((d) => d.source === 'slack')
				.forEach((d) => {
					checked[d.id] = true;
				});
		}
	})

	let channels = [{ label: 'All Channels', value: 'all' }];

	type Message = {
		id: string;
		avatar: string;
		username: string;
		content: string;
	};
	const channelMessages: { [id: string]: Message[] } = {
		all: [{ id: 'foo', avatar: mdiAccountSearch, username: 'Foo Bar', content: 'Lorem Ipsum' }]
	};
	const getMessages = (channelId: string) => {
		return channelMessages[channelId] || [];
	};
	let value = $state('all');

	const toggleInclusion = (message: Message) => {
		// dispatch('toggledata', { source: 'slack', id: message.id } as EventData);
	};
</script>

<Tabs
	options={channels}
	placement="top"
	bind:value
	classes={{
		tab: { root: 'bg-surface-300' },
		content: 'border'
	}}
>
	<div slot="content" let:value class="">
		{@const messages = getMessages(value)}
		{#each messages as msg, i}
			<ListItem classes={{ root: '!elevation-0 hover:bg-surface-300/50' }}>
				<div slot="avatar" class="flex flex-col">
					<Icon data={msg.avatar} />
					<span>time</span>
				</div>
				<span slot="title" class="text-sm text-neutral-50 block">
					{msg.username}
				</span>
				<span slot="subheading" class="text-sm">
					{msg.content}
				</span>
				<svelte:fragment slot="actions">
					<Checkbox
						checked={checked[msg.id]}
						on:change={() => {
							toggleInclusion(msg);
						}}
					/>
				</svelte:fragment>
			</ListItem>
		{/each}
	</div>
</Tabs>
