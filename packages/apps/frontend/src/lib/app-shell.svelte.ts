import { Context, watch, type Getter } from "runed";
import { onDestroy, type Component, type Snippet } from "svelte";
import type { ResolvedPathname } from "$app/types";

export type AppSidebarItem = {
	label: string;
	icon?: Component;
	href: string;
	subItems?: AppSidebarItem[];
};

export type AppSidebarGroup = {
	label?: string;
	items: AppSidebarItem[];
};

export type AppSidebarSearch = {
	placeholder: string;
};

export type AppSidebarModel = {
	isDefault?: boolean;
	search?: AppSidebarSearch;
	groups: AppSidebarGroup[];
};

export type PageBreadcrumb = {
	label: string;
	path: ResolvedPathname;
};

export type PageDescriptor = {
	title: string;
	status?: string;
	parents?: readonly PageBreadcrumb[];
	pageActions?: Snippet;
};

export class AppShellController {
	childSidebar = $state.raw<AppSidebarModel>();

	featureRail = $state(false);
	private featureRailOwner?: object;

	pageDescriptor = $state.raw<PageDescriptor>();
	private pageDescriptorOwner?: object;

	registerPageDescriptor(fn: Getter<PageDescriptor>) {
		const owner = {};
		this.pageDescriptorOwner = owner;
		this.pageDescriptor = fn();

		watch(fn, (descriptor) => {
			if (this.pageDescriptorOwner === owner) {
				this.pageDescriptor = descriptor;
			}
		});

		onDestroy(() => {
			if (this.pageDescriptorOwner === owner) {
				this.pageDescriptor = undefined;
				this.pageDescriptorOwner = undefined;
			}
		});
	}

	setChildSidebar(model: AppSidebarModel) {
		this.childSidebar = model;
	}

	clearChildSidebar() {
		this.childSidebar = undefined;
	}

	registerFeatureRail() {
		const owner = {};
		this.featureRailOwner = owner;
		this.featureRail = true;
		onDestroy(() => {
			if (this.featureRailOwner === owner) {
				this.featureRail = false;
				this.featureRailOwner = undefined;
			}
		});
	}
}

const ctx = new Context<AppShellController>("AppShellController");
export const initAppShell = () => ctx.set(new AppShellController());
export const useAppShell = () => ctx.get();

export const registerPageDescriptor = (getDescriptor: Getter<PageDescriptor>) => {
	useAppShell().registerPageDescriptor(getDescriptor);
};
