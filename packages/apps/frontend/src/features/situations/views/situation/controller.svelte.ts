import { Context, watch, type Getter } from "runed";
import { page } from "$app/state";
import { tick } from "svelte";
import { createQuery } from "@tanstack/svelte-query";
import { getSituationOptions } from "$lib/api";
import { resolve } from "$app/paths";

class SituationController {
	private id = $state<string>(null!);
	query = createQuery(() => getSituationOptions({ path: { id: this.id } }));

	constructor(idFn: Getter<string>) {
		watch(idFn, (id) => {
			this.id = id;
		});

		watch(
			() => [page.url.hash, this.query.data],
			() => {
				if (!page.url.hash.startsWith("#investigation-") || !this.query.data) return;
				const id = page.url.hash.slice(1);
				void tick().then(() => document.getElementById(id)?.scrollIntoView({ block: "start" }));
			}
		);
	}

	situation = $derived(this.query.data?.data);
	private mapHrefParams = $derived({ focus: this.situation?.attributes.knowledgeEntityId || "" });
	mapHref = $derived(`${resolve("/map")}?${new URLSearchParams(this.mapHrefParams)}`);
}

const ctx = new Context<SituationController>("SituationController");
export const initSituationController = (idFn: Getter<string>) => ctx.set(new SituationController(idFn));
