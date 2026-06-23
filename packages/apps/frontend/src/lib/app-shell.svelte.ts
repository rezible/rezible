import { Context, watch, type Getter } from "runed";
import type Avatar from "$components/common/entity-avatar/EntityAvatar.svelte";
import type { Component, ComponentProps } from "svelte";
import { page } from "$app/state";
import { afterNavigate } from "$app/navigation";
import type { Pathname } from "$app/types";
import type { RouteId } from "$app/types";
import type { SidebarModel } from "$features/app/components/app-shell/app-sidebar/controller.svelte";

export type PageBreadcrumb = {
	label?: string;
	path?: Pathname;
	avatar?: ComponentProps<typeof Avatar>;
};

export type PageActions<PComponent extends Component<any>> = {
	component: Component;
	propsFn?: () => ComponentProps<PComponent>;
	allowChildren: boolean;
	pathBase: string;
}

export class AppShellController {
	pageTitle = $state("Rezible")
	childSidebar = $state.raw<SidebarModel>();

	constructor() {
		afterNavigate(nav => {
			this.checkPageActions(nav.to?.route.id);
		});
	}

	pageActions = $state<PageActions<any>>();
	private checkPageActions(newRouteId?: RouteId | null) {
		if (!this.pageActions) return;
		const isChild = !!newRouteId && newRouteId.startsWith(this.pageActions.pathBase);
		if (!isChild || !this.pageActions.allowChildren) {this.pageActions = undefined}
	}

	setPageActions<PComponent extends Component<any>>(component: PComponent, allowChildren: boolean, propsFn?: () => ComponentProps<PComponent>) {
		this.pageActions = {component, allowChildren, propsFn, pathBase: $state.snapshot(page.route.id) ?? ""};
	}

	breadcrumbs = $state<PageBreadcrumb[]>([]);
	setPageBreadcrumbs(crumbsFn: Getter<PageBreadcrumb[]>) {
		watch(crumbsFn, crumbs => {this.breadcrumbs = crumbs});
	}

	setChildSidebar(model: SidebarModel) {
		this.childSidebar = model;
	}

	clearChildSidebar() {
		this.childSidebar = undefined;
	}
}

const ctx = new Context<AppShellController>("AppShellController");
export const initAppShell = () => ctx.set(new AppShellController());
export const useAppShell = () => ctx.get();

export const setPageBreadcrumbs = (crumbsFn: Getter<PageBreadcrumb[]>) => {
	useAppShell().setPageBreadcrumbs(crumbsFn);
}
