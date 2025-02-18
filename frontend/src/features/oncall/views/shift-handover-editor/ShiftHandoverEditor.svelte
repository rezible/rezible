<script lang="ts">
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { Button, Icon } from "svelte-ux";
	import { mdiSend, mdiPhoneForward } from "@mdi/js";
	import {
		getNextOncallShiftOptions,
		getOncallShiftHandoverOptions,
		getOncallShiftHandoverTemplateOptions,
		sendOncallShiftHandoverMutation,
		type OncallShift,
		type OncallShiftHandover,
		type OncallShiftHandoverTemplate,
		type SendOncallShiftHandoverAttributes,
	} from "$lib/api";
	import { getToastState } from "$components/toaster";
	import Avatar from "$components/avatar/Avatar.svelte";
	import ShiftAnnotationsList from "$features/oncall/components/shift-annotations/ShiftAnnotationsList.svelte";
	import { handoverState } from "./handover.svelte";
	import ReportEditor from "./ReportEditor.svelte";
	import ShiftReviewQuestionsDialog from "./ShiftReviewQuestionsDialog.svelte";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";

	type Props = { shift: OncallShift; handover?: OncallShiftHandover };
	const { shift, handover }: Props = $props();

	const shiftId = $derived(shift.id);

	const templateId = $derived(shift.attributes.roster.attributes.handoverTemplateId);
	const templateQuery = createQuery(() =>
		getOncallShiftHandoverTemplateOptions({ path: { id: templateId } })
	);

	const queryClient = useQueryClient();
	const nextShiftQuery = createQuery(() => getNextOncallShiftOptions({ path: { id: shiftId } }));
	const nextUser = $derived(nextShiftQuery.data?.data.attributes.user);

	const toastState = getToastState();

	let showReviewDialog = $state(false);
	const sendMutation = createMutation(() => ({
		...sendOncallShiftHandoverMutation(),
		onSuccess: () => {
			handoverState.setSent();
			toastState.add("Handover Sent", "Sent oncall shift handover", mdiPhoneForward);
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

<div class="px-3 flex-1 flex flex-col gap-2 max-h-full min-h-0 overflow-y-auto">
	<div class="flex gap-2">
		<div class="flex items-center gap-2 bg-neutral rounded w-fit h-full p-2 px-3 text-lg">
			<span class="">{handoverState.sent ? "Handed" : "Handing"} over to</span>
			<span class="font-bold flex gap-2 items-center">
				{nextUser?.attributes.name ?? ""}
				<Avatar kind="user" size={22} id={nextUser?.id ?? ""} />
			</span>
		</div>

		<div class="h-full w-0 border-l"></div>

		<Button
			size="lg"
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
					Send <Icon data={mdiSend} />
				{/if}
			</span>
		</Button>
	</div>

	<div class="grid grid-cols-5 gap-2 flex-1 min-h-0 max-h-full">
		<div
			class="col-span-2 flex flex-col gap-2 min-h-0 max-h-full overflow-y-auto border rounded-lg p-3 bg-surface-200"
		>
			<LoadingQueryWrapper query={templateQuery}>
				{#snippet view(template: OncallShiftHandoverTemplate)}
					<ReportEditor {shift} {template} {handover} />
				{/snippet}
			</LoadingQueryWrapper>
		</div>

		<div
			class="col-span-2 flex flex-col gap-2 min-h-0 border rounded-lg p-2 h-full max-h-full overflow-hidden"
		>
			<ShiftAnnotationsList editable={!handoverState.sent} {shiftId} />
		</div>
	</div>
</div>

{#if showReviewDialog}
	<ShiftReviewQuestionsDialog {shiftId} />
{/if}
