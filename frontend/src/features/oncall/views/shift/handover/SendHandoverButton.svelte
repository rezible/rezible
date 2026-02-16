<script lang="ts">
	import { createMutation } from "@tanstack/svelte-query";
	import { mdiSend, mdiPhoneForward } from "@mdi/js";
	import { sendOncallShiftHandoverMutation } from "$lib/api";
	import { useToastState } from "$lib/toasts.svelte";
	import { Button } from "$components/ui/button";
	import Icon from "$components/icon/Icon.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { useOncallShiftViewController } from "$features/oncall/views/shift";
	import { ShiftHandoverEditorState } from "$features/oncall/components/shift-handover-content/state.svelte";

	type Props = { 
		handoverState: ShiftHandoverEditorState;
		onSent: () => void;
	};
	const { handoverState, onSent }: Props = $props();

	const view = useOncallShiftViewController();

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
	color={handoverState.canSend ? "primary" : "default"}
	disabled={!handoverState.canSend || sendMutation.isPending}
	onclick={submitHandover}
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