import { getPlaybookOptions, updatePlaybookMutation, type UpdatePlaybookAttributes } from "$lib/api";
import { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class PlaybookViewState {
	playbookId = $state<string>(null!);
	editing = $state(false);

	private playbookQuery = createQuery(() => getPlaybookOptions({ path: { id: this.playbookId } }));
	playbook = $derived(this.playbookQuery.data?.data);
	playbookTitle = $derived(this.playbook?.attributes.title ?? "");
	playbookContent = $derived(this.playbook?.attributes.content);

	updatePlaybookMut = createMutation(() => ({
		...updatePlaybookMutation(),
		onSettled: () => {
			this.editing = false;
			this.editor?.setEditable(true);
		},
		onSuccess: (data) => {
			// TODO: optimistic update
			this.playbookQuery.refetch();
		}
	}));

	loading = $derived(this.updatePlaybookMut.isPending || this.playbookQuery.isLoading);

	editor = $state<SvelteEditor>();

	constructor(idFn: () => string) {
		this.playbookId = idFn();
		watch(idFn, id => {this.playbookId = id});
	}

	cancelEditing() {
		this.editing = false;
	}

	saveEdit() {
		this.editing = false;
		if (!this.editor || !this.playbook) return;
		this.editor.setEditable(false);

		const id = this.playbookId;
		const attributes: UpdatePlaybookAttributes = {
			title: this.playbook.attributes.title,
			content: this.editor.getHTML(),
		}
		this.updatePlaybookMut.mutateAsync({
			body: { attributes },
			path: { id },
		});
	}
}

export const playbookViewStateCtx = new Context<PlaybookViewState>("playbookView");