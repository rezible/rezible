import { getAlertOptions } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class AlertViewState {
	alertId = $state<string>(null!);

	private alertQuery = createQuery(() => getAlertOptions({ path: { id: this.alertId } }));
	alert = $derived(this.alertQuery.data?.data);
	alertTitle = $derived(this.alert?.attributes.title ?? "");
	
	constructor(idFn: () => string) {
		this.alertId = idFn();
		watch(idFn, id => {this.alertId = id});
	}
}

export const alertViewStateCtx = new Context<AlertViewState>("alertView");