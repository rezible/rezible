<script lang="ts">
	type Props = {
		dataValue: string;
	};
	let { dataValue = $bindable() }: Props = $props();

	let channels = [{ label: "#todo-incident-channel", value: "all" }];

	type Message = {
		id: string;
		username: string;
		content: string;
	};
	const channelMessages: { [id: string]: Message[] } = {
		all: [
			{
				id: "foo",
				username: "User Name",
				content: "message content",
			},
		],
	};
	const getMessages = (channelId: string) => {
		return channelMessages[channelId] || [];
	};
	let channelId = $state("all");

	const onSelected = (msgId: string) => {
		if (dataValue === msgId) {
			dataValue = "";
			return;
		}
		dataValue = $state.snapshot(msgId);
	};
</script>

<!-- <Tabs
	options={channels}
	placement="top"
	bind:value={channelId}
	classes={{
		tab: { root: "" },
		content: "border",
	}}
>
	<div slot="content" let:value class="">
		{@const messages = getMessages(value)}
		{#each messages as msg, i}
			{@const isSelected = dataValue === msg.id}
			<ListItem 
				classes={{ root: cls("!elevation-0", isSelected ? "bg-background/40" : "hover:bg-background/20") }} 
				onclick={e => {e.preventDefault(); onSelected(msg.id)}}
			>
				<div slot="avatar" class="flex flex-col">
					<svelte:component this={msg.avatar} aria-hidden="true" />
				</div>
				<span slot="title" class="text-sm text-neutral-50 block">
					{msg.username}
				</span>
				<span slot="subheading" class="text-sm">
					{msg.content}
				</span>
				<svelte:fragment slot="actions">
					<Checkbox circle checked={isSelected} size="lg" onchange={() => {onSelected(msg.id)}} />
				</svelte:fragment>
			</ListItem>
		{/each}
	</div>
</Tabs> -->
