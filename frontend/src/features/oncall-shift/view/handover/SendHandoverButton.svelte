<script lang="ts">
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { Button } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiSend, mdiPhoneForward } from "@mdi/js";
	import {
		getNextOncallShiftOptions,
		getOncallShiftHandoverOptions,
		sendOncallShiftHandoverMutation,
	} from "$lib/api";
	import { getToastState } from "$features/app-shell/lib/toasts.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { ShiftHandoverEditorState } from "$features/oncall-shift/components/shift-handover-content/state.svelte";

	type Props = { 
		handoverState: ShiftHandoverEditorState;
	};
	const { handoverState }: Props = $props();

	const shiftId = $derived(handoverState.handover?.attributes.shiftId || "");

	const nextShiftQuery = createQuery(() => ({
		...getNextOncallShiftOptions({ path: { id: shiftId } }),
		enabled: !!shiftId,
	}));
	const nextUser = $derived(nextShiftQuery.data?.data.attributes.user);

	const queryClient = useQueryClient();

	const toasts = getToastState();

	const sendMutation = createMutation(() => ({
		...sendOncallShiftHandoverMutation(),
		onSuccess: () => {
			// handoverState.setSent();
			toasts.add("Handover Sent", "Sent oncall shift handover", mdiPhoneForward);
			queryClient.invalidateQueries(getOncallShiftHandoverOptions({ path: { id: shiftId } }));
			// showReviewDialog;
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
	variant={handoverState.canSend ? "fill" : "fill-light"}
	color={handoverState.canSend ? "primary" : "default"}
	disabled={!handoverState.canSend}
	loading={sendMutation.isPending}
	on:click={submitHandover}
	classes={{ root: "gap-2 py-3" }}
>
	<span class="flex items-center gap-2 text-lg">
		{#if handoverState.isSent}
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