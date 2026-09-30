import { Context, watch, type Getter } from "runed";
import { onDestroy, type Component, type Snippet } from "svelte";
import type { ResolvedPathname } from "$app/types";
import type { StatusPresentation } from "$components/common/status-badge/status";

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

export type PageRelatedLink = {
	/** Stable key for keyed each blocks, e.g. the object ID. */
	key: string;
	/** Singular object kind, sentence case: "Incident", "Situation". */
	kind: string;
	icon: Component;
	label: string;
	path: ResolvedPathname;
	status?: StatusPresentation;
};

export type PageDescriptor = {
	title: string;
	/** The page object's primary status. Rendered once, in the header. */
	status?: StatusPresentation;
	parents?: readonly PageBreadcrumb[];
	/** Objects this page's object is linked to (situation ↔ incident). */
	related?: readonly PageRelatedLink[];
	/**
	 * When true, the view renders its own <h1> in content and the header's current
	 * breadcrumb is a non-heading element. Default false (header crumb is the <h1>).
	 */
	contentHeading?: boolean;
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
