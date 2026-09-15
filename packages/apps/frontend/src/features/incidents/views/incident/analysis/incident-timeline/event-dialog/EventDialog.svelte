<script lang="ts">
	import * as Dialog from "$components/ui/dialog";
	import { Button } from "$components/ui/button";
	import ErrorAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import EventAttributesEditor from "./EventAttributesEditor.svelte";
	import { useEventDialog } from "./controller.svelte";
	const editor = useEventDialog();
</script>

<Dialog.Root
	bind:open={
		() => editor.open,
		(open) => {
			if (!open) editor.clear();
		}
	}
>
	<Dialog.Content
		class="flex max-h-[calc(100dvh-2rem)] flex-col sm:max-w-xl"
		onCloseAutoFocus={(event) => {
			event.preventDefault();
			editor.restoreFocus();
		}}
		onInteractOutside={(event) => event.preventDefault()}
	>
		<Dialog.Header>
			<Dialog.Title>{editor.editingEntry ? "Edit" : "Create"} entry</Dialog.Title>
			<Dialog.Description>Update the narrative details of this analysis entry.</Dialog.Description>
		</Dialog.Header>
		<form
			class="flex min-h-0 flex-col gap-4"
			onsubmit={(event) => {
				event.preventDefault();
				void editor.confirm();
			}}
		>
			<div class="min-h-0 overflow-y-auto">
				<EventAttributesEditor />
			</div>
			{#if editor.error}
				<div role="alert">
					<ErrorAlert error={editor.error} />
				</div>
			{/if}
			<Dialog.Footer>
				<Button
					type="button"
					variant="outline"
					disabled={editor.loading}
					onclick={editor.clear}
				>Cancel</Button>
				<Button 
					type="submit" 
					disabled={!editor.title.trim() || editor.loading}
				>{editor.loading ? "Saving…" : "Save entry"}</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
