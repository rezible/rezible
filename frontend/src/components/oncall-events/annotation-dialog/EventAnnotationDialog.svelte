<script lang="ts">
	import { createMutation } from "@tanstack/svelte-query";
	import { mdiClose } from "@mdi/js";
	import { Button, Dialog } from "svelte-ux";
	import Header from "$components/header/Header.svelte";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { createOncallAnnotationMutation, updateOncallAnnotationMutation, type CreateOncallAnnotationRequestAttributes, type OncallAnnotation, type OncallAnnotationAlertFeedback, type OncallEvent, type OncallRoster } from "$src/lib/api";
	import EventAnnotationForm from "./EventAnnotationForm.svelte";
	import { attributesState } from "./attributes.svelte";
	
	type Props = {
		roster: OncallRoster;
		event?: OncallEvent;
		current?: OncallAnnotation;
		onClose: () => void;
	}
	const { roster, event, current, onClose }: Props = $props();

	const createMut = createMutation(() => ({
		...createOncallAnnotationMutation(),
		onSuccess: () => {
			onClose();
		}
	}));

	const updateMut = createMutation(() => ({
		...updateOncallAnnotationMutation(),
		onSuccess: () => {
			onClose();
		}
	}));

	const onConfirm = () => {
		if (!event) return;

		let alertFeedback: OncallAnnotationAlertFeedback | undefined = undefined;
		if (event?.attributes.kind === "alert") {
			alertFeedback = attributesState.getAlertFeedback();
		}
		const attributes = $state.snapshot({
			eventId: event.id,
			rosterId: roster.id,
			minutesOccupied: 0,
			notes: attributesState.notes,
			tags: attributesState.tags.values().toArray(),
			alertFeedback,
		})
		if (current) {
			updateMut.mutate({path: {id: current.id}, body: {attributes}});
		} else {
			createMut.mutate({body: {attributes}})
		}
	}

	const formAction = $derived(!!current ? "Update" : "Create");
</script>

<Dialog
	open={!!event}
	on:close={onClose}
	persistent
	portal
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-5xl my-2 min-h-0",
		root: "p-2",
	}}
>
	<div slot="header" class="border-b p-2" let:close>
		<Header title="{formAction} Annotation" subheading="For {roster.attributes.name}">
			{#snippet actions()}
				<Button on:click={() => close({ force: true })} iconOnly icon={mdiClose} />
			{/snippet}
		</Header>
	</div>

	<div slot="default" class="p-2 flex-1 min-h-0 max-h-full grid overflow-y-auto">
		{#if !!event}
			<EventAnnotationForm {event} {current} />
		{/if}
	</div>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			loading={false}
			closeText="Cancel"
			confirmText={formAction}
			{onClose}
			{onConfirm}
		/>
	</svelte:fragment>
</Dialog>
