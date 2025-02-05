<script lang="ts">
	import { mdiClose } from "@mdi/js";
	import { Button, Dialog, Header } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { analysis } from "../analysis.svelte";
	import ComponentAttributesEditor from "./ComponentAttributesEditor.svelte";
	import {
		createSystemComponentMutation,
		updateSystemComponentMutation,
		type CreateSystemComponentAttributes,
		type CreateSystemComponentData,
		type SystemComponent,
		type SystemComponentAttributes,
		type UpdateSystemComponentAttributes,
	} from "$src/lib/api";
	import { watch } from "runed";
	import ComponentSelector from "./ComponentSelector.svelte";
	import { createMutation } from "@tanstack/svelte-query";

	let purpose = $state<"add" | "create" | "edit">("add");
	const action = $derived.by(() => {
		switch (purpose) {
			case "add":
				return "Add";
			case "create":
				return "Create";
			case "edit":
				return "Edit";
		}
	});

	let selectedAddComponent = $state<SystemComponent>();

	const emptyComponentAttributes: SystemComponentAttributes = {
		constraints: [],
		controls: [],
		description: "",
		kind: "",
		name: "",
		signals: [],
		properties: {},
	};
	let attributes = $state<SystemComponentAttributes>(emptyComponentAttributes);

	const saveEnabled = $derived(true);

	watch(
		() => analysis.componentDialogOpen,
		(open) => {
			if (analysis.editingComponent) {
				purpose = "edit";
				attributes = $state.snapshot(analysis.editingComponent.attributes.component.attributes);
			}
			if (!open) {
				purpose = "add";
				selectedAddComponent = undefined;
				attributes = emptyComponentAttributes;
			}
		}
	);

	const doClose = () => {
		// go back from create screen
		if (purpose === "create") {
			purpose = "add";
			return;
		}
		analysis.setComponentDialogOpen(false);
	};

	const mutationCallbacks = {
		onSuccess: doClose,
	};

	const createComponentMutation = createMutation(() => ({
		...createSystemComponentMutation(),
		...mutationCallbacks,
	}));

	const componentId = $derived(analysis.editingComponent?.attributes.component.id ?? "");
	const updateComponentMutation = createMutation(() => ({
		...updateSystemComponentMutation(),
		...mutationCallbacks,
	}));

	const loading = $derived(createComponentMutation.isPending);

	const onConfirm = () => {
		if (purpose === "create") {
			const reqAttributes: CreateSystemComponentAttributes = {
				name: attributes.name,
			};
			createComponentMutation.mutate({ body: { attributes: reqAttributes } });
		} else if (purpose === "edit") {
			const reqAttributes: UpdateSystemComponentAttributes = {
				name: attributes.name,
			};
			updateComponentMutation.mutate({
				path: { id: componentId },
				body: { attributes: reqAttributes },
			});
		} else if (purpose === "add") {
		}
	};
</script>

<Dialog
	open={analysis.componentDialogOpen}
	on:close={() => {
		analysis.setComponentDialogOpen(false);
	}}
	persistent
	portal
	{loading}
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-7xl my-2 min-h-0",
		root: "p-2",
	}}
>
	<div slot="header" class="border-b p-2" let:close>
		<Header title="{action} System Component">
			<svelte:fragment slot="actions">
				<Button on:click={() => close({ force: true })} iconOnly icon={mdiClose} />
			</svelte:fragment>
		</Header>
	</div>

	<div slot="default" class="p-2 flex-1 min-h-0 max-h-full grid">
		{#if analysis.componentDialogOpen}
			{#if purpose === "add"}
				<ComponentSelector
					bind:selected={selectedAddComponent}
					onCreateNew={() => {
						purpose = "create";
					}}
				/>
			{:else}
				<ComponentAttributesEditor
					bind:name={attributes.name}
					bind:description={attributes.description}
					bind:kind={attributes.kind}
					bind:constraints={attributes.constraints}
					bind:signals={attributes.signals}
					bind:controls={attributes.controls}
				/>
			{/if}
		{/if}
	</div>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			{loading}
			closeText="Cancel"
			confirmText={purpose == "edit" ? "Save" : action}
			onClose={doClose}
			{onConfirm}
			{saveEnabled}
		/>
	</svelte:fragment>
</Dialog>
