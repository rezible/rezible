import { Context, type Getter } from "runed";
import type { SystemViewParam } from "$params/systemView";

const screenTitle = (view: SystemViewParam): string => {
	switch (view) {
		case "catalogue":
			return "Catalogue";
		case "saved-views":
			return "Saved Views";
		default:
			return "Map";
	}
};

export class SystemViewController {
	private currentView: () => SystemViewParam = () => undefined;
	title = $derived(screenTitle(this.currentView()));

	constructor(currentView: () => SystemViewParam) {
		this.currentView = currentView;
	}
}

const ctx = new Context<SystemViewController>("SystemViewController");
export const initSystemViewController = (viewFn: Getter<SystemViewParam>) => ctx.set(new SystemViewController(viewFn));
export const useSystemViewController = () => ctx.get();
