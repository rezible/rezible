<script lang="ts">
	import * as Dialog from "$components/ui/dialog";
	import { Spinner } from "$components/ui/spinner";
	import { useIncidentCreateDialog } from "./controller.svelte";
	import CreateIncidentForm from "./CreateIncidentForm.svelte";

	const controller = useIncidentCreateDialog();
</script>

<Dialog.Root bind:open={() => controller.open, controller.setOpen}>
	<Dialog.Content
		class="max-h-[calc(100dvh-2rem)] max-w-2xl overflow-hidden p-0 sm:max-w-2xl"
		showCloseButton
	>
		<div class="border-b px-4 py-3">
			<Dialog.Header>
				<Dialog.Title>Create Incident</Dialog.Title>
				<Dialog.Description>Open a new incident</Dialog.Description>
			</Dialog.Header>
		</div>

		{#if controller.loading}
			<div class="w-full h-52 grid place-items-center">
				<Spinner />
			</div>
		{:else}
			<CreateIncidentForm />
		{/if}
	</Dialog.Content>
</Dialog.Root>
