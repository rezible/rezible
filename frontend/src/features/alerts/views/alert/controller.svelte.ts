import { getAlertOptions } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class AlertViewController {
	alertId = $state<string>(null!);

	private alertQuery = createQuery(() => getAlertOptions({ path: { id: this.alertId } }));
	alert = $derived(this.alertQuery.data?.data);
	alertTitle = $derived(this.alert?.attributes.title ?? "");
	
	constructor(idFn: Getter<string>) {
		this.alertId = idFn();
		watch(idFn, id => {this.alertId = id});
	}
}

const ctx = new Context<AlertViewController>("alertView");
export const initAlertViewController = (idFn: Getter<string>) => ctx.set(new AlertViewController(idFn));
export const useAlertViewController = () => ctx.get();