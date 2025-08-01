import { getAlertOptions } from "$lib/api";
import type { IdFunc } from "$lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class AlertViewState {
	alertId = $state<string>(null!);

	private alertQuery = createQuery(() => getAlertOptions({ path: { id: this.alertId } }));
	alert = $derived(this.alertQuery.data?.data);
	alertTitle = $derived(this.alert?.attributes.title ?? "");
	
	constructor(idFn: IdFunc) {
		this.alertId = idFn();
		watch(idFn, id => {this.alertId = id});
	}
}

const alertViewStateCtx = new Context<AlertViewState>("alertView");
export const setAlertViewState = (idFn: IdFunc) => alertViewStateCtx.set(new AlertViewState(idFn));
export const useAlertViewState = () => alertViewStateCtx.get();