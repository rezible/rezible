import { Context, watch, type Getter } from "runed";
import { onDestroy, type Component, type ComponentProps } from "svelte";
import type { ResolvedPathname } from "$app/types";

type AnyComponent = Component<any>;

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

export type PageDescriptor<C extends AnyComponent = Component> = {
	title: string;
	parents?: readonly PageBreadcrumb[];
	actions?: {
		component: C;
		props?: ComponentProps<C>;
	};
};

export class AppShellController {
	childSidebar = $state.raw<AppSidebarModel>();
	pageDescriptor = $state.raw<PageDescriptor<AnyComponent>>();
	private pageDescriptorOwner?: object;

	registerPageDescriptor<C extends AnyComponent>(getDescriptor: Getter<PageDescriptor<C>>) {
		const owner = {};
		this.pageDescriptorOwner = owner;
		this.pageDescriptor = getDescriptor();

		watch(getDescriptor, (descriptor) => {
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
}

const ctx = new Context<AppShellController>("AppShellController");
export const initAppShell = () => ctx.set(new AppShellController());
export const useAppShell = () => ctx.get();

export const registerPageDescriptor = <C extends AnyComponent>(getDescriptor: Getter<PageDescriptor<C>>) => {
	useAppShell().registerPageDescriptor(getDescriptor);
};
