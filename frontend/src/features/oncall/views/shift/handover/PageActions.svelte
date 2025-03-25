<script lang="ts">
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import {
		getNextOncallShiftOptions,
		getOncallShiftHandoverOptions,
		getOncallShiftOptions,
		sendOncallShiftHandoverMutation,
	} from "$lib/api";
	import { Button, Icon } from "svelte-ux";
	import { mdiSend, mdiPhoneForward } from "@mdi/js";
	import { getToastState } from "$features/app/lib/toasts.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { handoverState } from "./handover.svelte";
	import { page } from "$app/state";


	const shiftId = $derived(page.params.id);
	// const query = createQuery(() => getOncallShiftOptions({ path: { id: shiftId } }));
	// const shift = $derived(query.data?.data);

	const queryClient = useQueryClient();
	const nextShiftQuery = createQuery(() => getNextOncallShiftOptions({ path: { id: shiftId } }));
	const nextUser = $derived(nextShiftQuery.data?.data.attributes.user);

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
		const content = handoverState.getSectionContent();
		console.log(content);
		sendMutation.mutate({
			path: { id: shiftId },
			body: { attributes: { content } },
		});
	};

	const canSend = $derived(!handoverState.sent && !handoverState.isEmpty);
</script>

<Button
	variant={canSend ? "fill" : "fill-light"}
	color={canSend ? "primary" : "default"}
	disabled={!canSend}
	loading={sendMutation.isPending}
	on:click={submitHandover}
	classes={{ root: "gap-2" }}
>
	<span class="flex items-center gap-3 text-lg">
		{#if handoverState.sent}
			Sent
		{:else}
			Send Handover to
			<span class="font-bold flex gap-2 items-center">
				{nextUser?.attributes.name ?? ""}
				<Avatar kind="user" size={22} id={nextUser?.id ?? ""} />
			</span>
			<Icon data={mdiSend} />
		{/if}
	</span>
</Button>