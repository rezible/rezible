<script lang="ts">
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { Button, Icon } from "svelte-ux";
	import { mdiSend, mdiPhoneForward } from "@mdi/js";
	import {
		getNextOncallShiftOptions,
		getOncallShiftHandoverOptions,
		sendOncallShiftHandoverMutation,
	} from "$lib/api";
	import { getToastState } from "$features/app/lib/toasts.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { HandoverEditorState } from "./state.svelte";

	type Props = { 
		shiftId: string;
		handoverState: HandoverEditorState;
	};
	const { shiftId, handoverState }: Props = $props();

	const nextShiftQuery = createQuery(() => getNextOncallShiftOptions({ path: { id: shiftId } }));
	const nextUser = $derived(nextShiftQuery.data?.data.attributes.user);

	const queryClient = useQueryClient();
	const canSend = $derived(!handoverState.sent && !handoverState.isEmpty);

	const toasts = getToastState();

	const sendMutation = createMutation(() => ({
		...sendOncallShiftHandoverMutation(),
		onSuccess: () => {
			handoverState.setSent();
			toasts.add("Handover Sent", "Sent oncall shift handover", mdiPhoneForward);
			queryClient.invalidateQueries(getOncallShiftHandoverOptions({ path: { id: shiftId } }));
			// showReviewDialog = true;
		},
	}));

	const submitHandover = () => {
		sendMutation.mutate({
			path: { id: shiftId },
			body: { attributes: {  } },
		});
	};
</script>

<Button
	variant={canSend ? "fill" : "fill-light"}
	color={canSend ? "primary" : "default"}
	disabled={!canSend}
	loading={sendMutation.isPending}
	on:click={submitHandover}
	classes={{ root: "gap-2 py-3" }}
>
	<span class="flex items-center gap-2 text-lg">
		{#if handoverState.sent}
			Sent
		{:else}
			Send to
			<span class="font-bold flex gap-2 items-center">
				{nextUser?.attributes.name ?? ""}
				<Avatar kind="user" size={22} id={nextUser?.id ?? ""} />
			</span>
			<Icon data={mdiSend} />
		{/if}
	</span>
</Button>