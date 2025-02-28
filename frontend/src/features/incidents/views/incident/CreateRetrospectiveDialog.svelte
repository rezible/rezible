<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import { watch } from "runed";
	import { Dialog, Header } from "svelte-ux";
	import { createRetrospectiveMutation } from "$lib/api";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { createMutation, useQueryClient, type QueryKey } from "@tanstack/svelte-query";

	type Props = {
		open: boolean;
		incidentId: string;
		isIncidentView: boolean;
		queryKey: QueryKey;
	};
	let { open = $bindable(), incidentId, isIncidentView, queryKey }: Props = $props();

	const queryClient = useQueryClient();

	const createRetroMut = createMutation(() => ({
		...createRetrospectiveMutation(),
		onSuccess: (resp) => {
			queryClient.setQueryData(queryKey, resp);
			open = false;
		},
	}));
	const onConfirmCreateRetrospective = () => {
		createRetroMut.mutate({ body: { attributes: { incidentId, systemAnalysis: true } } });
	};

	const onCloseRetroDialog = () => {
		open = false;
		if (!isIncidentView) goto(`/incidents/${incidentId}`);
	};

	const isRetroView = $derived(!isIncidentView);
	watch(
		() => isRetroView,
		(forceOpen) => {
			open = open || forceOpen;
		}
	);
</script>

<Dialog
	{open}
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
