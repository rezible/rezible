<script lang="ts">
	import { createMutation } from "@tanstack/svelte-query";
	import { goto } from "$app/navigation";
	import { watch } from "runed";
	import { createRetrospectiveMutation } from "$lib/api";
	import { useIncidentViewState } from "./viewState.svelte";

	import { Dialog } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import Header from "$components/header/Header.svelte";

	type Props = {
		isIncidentView: boolean;
	};
	let { isIncidentView }: Props = $props();

	const viewState = useIncidentViewState();
	const incidentId = $derived(viewState.incident?.id);

	const createRetroMut = createMutation(() => ({
		...createRetrospectiveMutation(),
		onSuccess: (resp) => {
			viewState.onRetrospectiveCreated(resp.data);
		},
	}));
	const onConfirmCreateRetrospective = () => {
		if (!incidentId) return;
		createRetroMut.mutate({ body: { attributes: { incidentId, systemAnalysis: true } } });
	};

	const onCloseRetroDialog = () => {
		viewState.createRetrospectiveDialogOpen = false;
		if (!isIncidentView) goto(`/incidents/${incidentId}`);
	};

	watch(
		() => isIncidentView,
		(isIncidentView) => {
			const shouldBeOpen = !isIncidentView || viewState.createRetrospectiveDialogOpen;
			viewState.createRetrospectiveDialogOpen = shouldBeOpen;
		}
	);
</script>

<Dialog
	open={viewState.createRetrospectiveDialogOpen}
	persistent
	portal
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-7xl my-2",
		root: "p-2",
	}}
>
	<div slot="header" class="border-b p-2">
		<Header title="Create Incident Retrospective" />
	</div>

	<svelte:fragment slot="default">
		<span>todo: allow specifying retro options</span>
	</svelte:fragment>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			loading={createRetroMut.isPending}
			closeText="Cancel"
			confirmText="Create"
			onClose={onCloseRetroDialog}
			onConfirm={onConfirmCreateRetrospective}
		/>
	</svelte:fragment>
</Dialog>
