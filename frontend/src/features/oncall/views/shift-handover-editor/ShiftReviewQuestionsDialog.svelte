<script lang="ts">
	import { createOncallShiftAnnotationMutation } from "$lib/api";
	import { Icon, Button, TextField, Header, Dialog } from "svelte-ux";
	import { createMutation, createQuery } from "@tanstack/svelte-query";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { goto } from "$app/navigation";

	type Props = {
		shiftId: string;
	};
	let { shiftId }: Props = $props();

	const reviewShiftMutation = createMutation(() => ({
		...createOncallShiftAnnotationMutation(),
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
