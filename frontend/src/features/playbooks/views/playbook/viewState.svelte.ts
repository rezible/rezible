import { getPlaybookOptions } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class PlaybookViewState {
	playbookId = $state<string>(null!);
	editing = $state(false);

	private playbookQuery = createQuery(() => getPlaybookOptions({ path: { id: this.playbookId } }));
	playbook = $derived(this.playbookQuery.data?.data);
	playbookTitle = $derived(this.playbook?.attributes.title ?? "");
	playbookContent = $derived(this.playbook?.attributes.content);

	loading = $derived(this.playbookQuery.isLoading);

	constructor(idFn: () => string) {
		this.playbookId = idFn();
		watch(idFn, id => {this.playbookId = id});
	}

	cancelEditing() {
		this.editing = false;
	}

	saveEdit() {
		this.editing = false;
	}
}

export const playbookViewStateCtx = new Context<PlaybookViewState>("playbookView");