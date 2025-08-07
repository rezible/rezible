import { getAlertOptions, listOncallEventsOptions, type ListOncallEventsData } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { getLocalTimeZone, now } from "@internationalized/date";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

const to = now(getLocalTimeZone()).toAbsoluteString();
const from = now(getLocalTimeZone()).subtract({ days: 7 }).toAbsoluteString();

export class AlertViewState {
	alertId = $state<string>(null!);

	private alertQuery = createQuery(() => getAlertOptions({ path: { id: this.alertId } }));
	alert = $derived(this.alertQuery.data?.data);
	alertTitle = $derived(this.alert?.attributes.title ?? "");
	
	constructor(idFn: Getter<string>) {
		this.alertId = idFn();
		watch(idFn, id => {this.alertId = id});
	}

	private eventsQueryData = $derived<ListOncallEventsData["query"]>({ from, to, alertId: this.alertId })
	private eventsQuery = createQuery(() => listOncallEventsOptions({ query: this.eventsQueryData }));
	events = $derived(this.eventsQuery.data?.data);
}

const alertViewStateCtx = new Context<AlertViewState>("alertView");
export const setAlertViewState = (idFn: Getter<string>) => alertViewStateCtx.set(new AlertViewState(idFn));
export const useAlertViewState = () => alertViewStateCtx.get();