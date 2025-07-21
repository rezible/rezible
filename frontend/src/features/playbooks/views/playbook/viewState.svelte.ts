import { getPlaybookOptions } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class PlaybookViewState {
	playbookId = $state<string>(null!);
	constructor(idFn: () => string) {
		this.playbookId = idFn();
		watch(idFn, id => {this.playbookId = id});
	}

	private playbookQuery = createQuery(() => getPlaybookOptions({ path: { id: this.playbookId } }));
	playbook = $derived(this.playbookQuery.data?.data);
	playbookTitle = $derived(this.playbook?.attributes.title ?? "");
}

export const playbookViewStateCtx = new Context<PlaybookViewState>("playbookView");