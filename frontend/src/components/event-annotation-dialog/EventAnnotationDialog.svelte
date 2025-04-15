<script lang="ts">
	import { mdiClose } from "@mdi/js";
	import { Button, Dialog, Header } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { createOncallAnnotationMutation, type CreateOncallAnnotationRequestBody, type OncallAnnotation, type OncallEvent } from "$src/lib/api";
	import EventAnnotationForm from "./EventAnnotationForm.svelte";
	import { createMutation } from "@tanstack/svelte-query";
	
	type Props = {
		event?: OncallEvent;
		current?: OncallAnnotation;
		onClose: () => void;
	}
	const { event, current, onClose }: Props = $props();

	const createMut = createMutation(() => ({
		...createOncallAnnotationMutation(),
		onSuccess: () => {
			onClose();
		}
	}));

	let mutationData = $state<CreateOncallAnnotationRequestBody>()

	const onConfirm = () => {
		createMut.mutate
	}
</script>

<Dialog
	open={!!event}
	on:close={onClose}
	persistent
	portal
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-7xl my-2 min-h-0",
		root: "p-2",
	}}
>
	<div slot="header" class="border-b p-2" let:close>
		<Header title="Annotate Event">
			<svelte:fragment slot="actions">
				<Button on:click={() => close({ force: true })} iconOnly icon={mdiClose} />
			</svelte:fragment>
		</Header>
	</div>

	<div slot="default" class="p-2 flex-1 min-h-0 max-h-full grid">
		{#if !!event}
			<EventAnnotationForm {event} />
		{/if}
	</div>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			loading={false}
			closeText="Cancel"
			confirmText="Save"
			saveEnabled={false}
			{onClose}
			{onConfirm}
		/>
	</svelte:fragment>
</Dialog>
