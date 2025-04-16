<script lang="ts">
	import { createOncallAnnotationMutation } from "$lib/api";
	import { Header, Dialog } from "svelte-ux";
	import { createMutation } from "@tanstack/svelte-query";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { goto } from "$app/navigation";
	
	import { shiftViewStateCtx } from "../context.svelte";

	const viewState = shiftViewStateCtx.get();
	const shiftId = $derived(viewState.shiftId);

	const reviewShiftMutation = createMutation(() => ({
		// TODO: use correct query
		...createOncallAnnotationMutation(),
		onSuccess: () => {
			goto("/oncall/shifts/" + shiftId);
		},
	}));

	const skipReview = () => {};
	const saveReview = () => {};
</script>

<Dialog
	open
	loading={reviewShiftMutation.isPending}
	persistent
	portal
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-7xl my-2",
		root: "p-4",
	}}
>
	<div slot="header" class="border-b p-2" let:close>
		<Header title="Reviewing Shift" />
	</div>

	<svelte:fragment>
		{@render dialogBody()}
	</svelte:fragment>

	<div slot="actions">
		<ConfirmButtons closeText="Skip" onClose={skipReview} onConfirm={saveReview} confirmText="Submit" />
	</div>
</Dialog>

{#snippet dialogBody()}
	<div class="flex flex-col gap-2 overflow-y-auto p-2">
		<span>review questions</span>
	</div>
{/snippet}
