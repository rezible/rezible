import { getAlertOptions, getOncallRosterOptions, getPlaybookOptions } from "$src/lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class AlertViewState {
	alertId = $state<string>(null!);
	constructor(idFn: () => string) {
		this.alertId = idFn();
		watch(idFn, id => {this.alertId = id});
	}

	private alertQuery = createQuery(() => getAlertOptions({ path: { id: this.alertId } }));
	alert = $derived(this.alertQuery.data?.data);
	alertTitle = $derived(this.alert?.attributes.title ?? "");
}

export const alertViewStateCtx = new Context<AlertViewState>("alertView");