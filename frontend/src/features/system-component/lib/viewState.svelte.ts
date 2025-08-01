import { getSystemComponentOptions } from "$lib/api";
import type { Getter } from "$src/lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

class SystemComponentViewState {
	componentId = $state<string>(null!);

	constructor(idFn: Getter<string>) {
		this.componentId = idFn();
		watch(idFn, id => {this.componentId = id});
	}

	private query = createQuery(() => ({
		...getSystemComponentOptions({ path: { id: this.componentId } }),
	}));

	component = $derived(this.query.data?.data);
	componentName = $derived(this.component?.attributes.name ?? "");
}

const ctx = new Context<SystemComponentViewState>("SystemComponentView");
export const setSystemComponentViewState = (idFn: Getter<string>) => ctx.set(new SystemComponentViewState(idFn));
export const useSystemComponentViewState = () => ctx.get();