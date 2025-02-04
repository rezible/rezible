<script lang="ts">
	import { Button, Checkbox, Icon, ListItem, Tabs } from "svelte-ux";
	import { mdiAccount, mdiLinkPlus, mdiPlus } from "@mdi/js";

	type Props = {
		onLinked: (id: string) => void;
	};
	const { onLinked }: Props = $props();

	let channels = [{ label: "#todo-incident-channel", value: "all" }];

	type Message = {
		id: string;
		avatar: string;
		username: string;
		content: string;
	};
	const channelMessages: { [id: string]: Message[] } = {
		all: [
			{
				id: "foo",
				avatar: mdiAccount,
				username: "User Name",
				content: "message content",
			},
		],
	};
	const getMessages = (channelId: string) => {
		return channelMessages[channelId] || [];
	};
	let channelId = $state("all");
</script>

<Tabs
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
			<ListItem
				classes={{ root: "!elevation-0 hover:bg-surface-300/20" }}
			>
				<div slot="avatar" class="flex flex-col">
					<Icon data={msg.avatar} />
				</div>
				<span slot="title" class="text-sm text-neutral-50 block">
					{msg.username}
				</span>
				<span slot="subheading" class="text-sm">
					{msg.content}
				</span>
				<svelte:fragment slot="actions">
					<Button
						on:click={() => {
							onLinked(msg.id);
						}}
					>
						<span class="flex items-center gap-2">
							Add
							<Icon data={mdiLinkPlus} />
						</span>
					</Button>
				</svelte:fragment>
			</ListItem>
		{/each}
	</div>
</Tabs>
