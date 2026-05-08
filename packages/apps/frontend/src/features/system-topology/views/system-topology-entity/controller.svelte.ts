import { getSystemTopologyEntityOptions } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";

class SystemTopologyEntityViewController {
	entityId = $state<string>(null!);

	constructor(idFn: Getter<string>) {
		this.entityId = idFn();
		watch(idFn, id => {this.entityId = id});
	}

	private query = createQuery(() => ({
		...getSystemTopologyEntityOptions({ path: { id: this.entityId } }),
	}));

	entity = $derived(this.query.data?.data);
	entityName = $derived(this.entity?.attributes.displayName ?? "");
}

const ctx = new Context<SystemTopologyEntityViewController>("SystemTopologyEntityViewController");
export const initSystemTopologyEntityViewController = (idFn: Getter<string>) => ctx.set(new SystemTopologyEntityViewController(idFn));
export const useSystemTopologyEntityViewController = () => ctx.get();
