import { getAlertOptions } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class AlertViewState {
	alertId = $state<string>(null!);

	private alertQuery = createQuery(() => getAlertOptions({ path: { id: this.alertId } }));
	alert = $derived(this.alertQuery.data?.data);
	alertTitle = $derived(this.alert?.attributes.title ?? "");
	
	constructor(idFn: Getter<string>) {
		this.alertId = idFn();
		watch(idFn, id => {this.alertId = id});
	}
}

const alertViewStateCtx = new Context<AlertViewState>("alertView");
export const setAlertViewState = (idFn: Getter<string>) => alertViewStateCtx.set(new AlertViewState(idFn));
export const useAlertViewState = () => alertViewStateCtx.get();