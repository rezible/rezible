<script lang="ts">
	import { createQuery, queryOptions, useQueryClient, type QueryState } from '@tanstack/svelte-query';
	import { listDebriefMessagesOptions, type Incident, type IncidentDebrief, type IncidentDebriefMessage, type ListIncidentDebriefMessagesResponseBody } from '$lib/api';
	import GetStarted from './GetStarted.svelte';
	import MessageEntryBox from './MessageEntryBox.svelte';
	import MessagesView from './MessagesView.svelte';

	type Props = { 
		debrief: IncidentDebrief;
	};
	const { debrief }: Props = $props();

	const queryClient = useQueryClient();

	const getLatestMessage = (messages: IncidentDebriefMessage[]) => {
		if (messages.length === 0) return null;
		let mostRecent = messages[0];
		let createdAt = Date.parse(mostRecent.attributes.createdAt);
		for (let i = 0; i < messages.length; i++) {
			const msg = messages[i];
			const msgCreatedAt = Date.parse(msg.attributes.createdAt);
			if (msgCreatedAt > createdAt) {
				mostRecent = msg;
				createdAt = msgCreatedAt;
			}
		}
		return mostRecent;
	}
	const sortMessages = (messages: IncidentDebriefMessage[]) => {
		return messages.toSorted((a, b) => {
			const older = Date.parse(a.attributes.createdAt) < Date.parse(b.attributes.createdAt);
			return older ? -1 : 1;
		})
	}

	const isUserMessage = (msg: IncidentDebriefMessage) => msg.attributes.type === "user";
	
	const REFETCH_INTEVAL_MS = 1000;
	const shouldQueryPoll = (state: QueryState<ListIncidentDebriefMessagesResponseBody, Error>) => {
		if (state.error) return false;

		const latestMessage = getLatestMessage(state.data?.data ?? []);
		return !latestMessage || isUserMessage(latestMessage);
	}

	const started = $derived(debrief.attributes.started);
	const messagesQueryOptions = $derived(queryOptions({
		...listDebriefMessagesOptions({path: {id: debrief.id}}),
		select: res => sortMessages(res.data),
		refetchInterval: ({state}) => (shouldQueryPoll(state) ? REFETCH_INTEVAL_MS : false),
		enabled: started,
	}));
	const queryKey = $derived(messagesQueryOptions.queryKey);
	const messagesQuery = createQuery(() => messagesQueryOptions);

	const messages = $derived(messagesQuery.data ?? []);
	const lastMessage = $derived(messages.length > 0 ? messages[messages.length - 1] : undefined);
	const waitingForResponse = $derived(!lastMessage || isUserMessage(lastMessage));

	const onMessageAdded = (msg: IncidentDebriefMessage) => 
		queryClient.setQueryData(queryKey, curData => {
			if (!curData) return curData;
			const newData = structuredClone(curData);
			newData.data = [...curData.data, msg];
			return newData;
		});
</script>

<div class="flex flex-col overflow-y-auto gap-2 border p-2">
	{#if started}
		<div class="grow block overflow-y-auto pb-1">
			{#if messagesQuery.isError}
				<span>query error: {messagesQuery.error}</span>
			{:else}
				<MessagesView 
					{messages}
					{waitingForResponse}
				/>
			{/if}
		</div>

		<div class="h-fit w-full rounded-xl">
			<MessageEntryBox
				{debrief}
				{lastMessage}
				disabled={waitingForResponse || messagesQuery.isError}
				onAdded={onMessageAdded}
			/>
		</div>
	{:else}
		<GetStarted {debrief} />
	{/if}
</div>