<script lang="ts">
	import { createMutation } from "@tanstack/svelte-query";
	import Button from "$components/button/Button.svelte";
	import { mdiSend, mdiPhoneForward } from "@mdi/js";
	import { sendOncallShiftHandoverMutation } from "$lib/api";
	import { useToastState } from "$lib/toasts.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { useOncallShiftViewState } from "$features/oncall-shift";
	import { ShiftHandoverEditorState } from "$features/oncall-shift/components/shift-handover-content/state.svelte";

	type Props = { 
		handoverState: ShiftHandoverEditorState;
		onSent: () => void;
	};
	const { handoverState, onSent }: Props = $props();

	const view = useOncallShiftViewState();

	const nextUser = $derived(view.nextShift?.attributes.user);

	const toasts = useToastState();

	const sendMutation = createMutation(() => ({
		...sendOncallShiftHandoverMutation(),
		onSuccess: () => {
			// handoverState.setSent();
			toasts.add("Handover Sent", "Sent oncall shift handover", mdiPhoneForward);
			// showReviewDialog;
			onSent();
		},
	}));

	const submitHandover = () => {
		sendMutation.mutate({
			path: { id: view.shiftId },
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