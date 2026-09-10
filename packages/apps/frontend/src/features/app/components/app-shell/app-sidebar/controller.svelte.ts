import { Context } from "runed";
import { page } from "$app/state";
import { useSidebar } from "$components/ui/sidebar";
import { useUserSessionState } from "$lib/user-session.svelte";
import {
	useAppShell,
	type AppSidebarGroup,
	type AppSidebarItem,
	type AppSidebarModel,
} from "$lib/app-shell.svelte";

import RiRadarLine from "remixicon-svelte/icons/radar-line";
import RiPulseLine from "remixicon-svelte/icons/pulse-line";
import RiConnectorLine from "remixicon-svelte/icons/connector-line";

const isActive = (href: string, pathname: string) => {
	if (href === "/" || pathname === "/") return pathname === href;
	return pathname === href || pathname.startsWith(`${href}/`);
};

const makeActiveStatus = (pathname: string, groups: AppSidebarGroup[]) => {
	let deepestActiveItem = "";
	let activeSubItems = new Map<string, Set<string>>();

	const checkItemActive = (item: AppSidebarItem) => {
		if (item.href.length > deepestActiveItem.length && isActive(item.href, pathname)) {
			deepestActiveItem = item.href;
		}
		const subActive = new Set<string>();
		item.subItems?.forEach((sub) => {
			if (isActive(sub.href, pathname)) subActive.add(sub.href);
		});
		activeSubItems.set(item.href, subActive);
	};

	groups.forEach((g) => g.items.forEach(checkItemActive));

	return { deepestActiveItem, activeSubItems };
};

const filterGroups = (groups: AppSidebarGroup[] | undefined, query?: string) => {
	if (!groups) return [];
	if (!query) return groups;
	return groups
		.map((group) => {
			const labelMatches = group.label?.toLowerCase().includes(query);
			const items = group.items.filter(
				(item) =>
					labelMatches ||
					item.label.toLowerCase().includes(query) ||
					item.subItems?.some((subItem) => subItem.label.toLowerCase().includes(query))
			);
			return { ...group, items };
		})
		.filter((group) => group.items.length > 0);
};

const defaultSidebarModel: AppSidebarModel = {
	groups: [
		{
			label: "General",
			items: [
				{ label: "Home", href: "/", icon: RiRadarLine },
				{ label: "Incidents", href: "/incidents", icon: RiPulseLine },
				{ label: "Situations", href: "/situations", icon: RiRadarLine },
				{ label: "Signals", href: "/signals", icon: RiPulseLine },
				{ label: "System Map", href: "/map", icon: RiConnectorLine },
			],
		},
	],
};

class AppSidebarController {
	private shell = useAppShell();

	private session = useUserSessionState();
	preloadHome = $derived<"tap" | "hover">(this.session.error ? "tap" : "hover");

	private sidebarState = useSidebar();

	isOpen = $derived(this.sidebarState.effectiveOpen);
	collapsed = $derived(this.sidebarState.state === "collapsed");

	isDefault = $derived(!this.shell.childSidebar);
	model = $derived(!this.shell.childSidebar ? defaultSidebarModel : this.shell.childSidebar);

	showUserMenu = $derived(!!this.isDefault);

	showSearch = $derived(!!this.model.search);
	searchQuery = $state("");
	private normalizedQuery = $derived(this.searchQuery.trim().toLowerCase());

	groups = $derived(
		this.showSearch ? filterGroups(this.model.groups, this.normalizedQuery) : this.model.groups
	);

	activeStatus = $derived(makeActiveStatus(page.url.pathname, this.groups));
}

const ctx = new Context<AppSidebarController>("AppSidebarController");
export const initAppSidebarController = () => ctx.set(new AppSidebarController());
export const useAppSidebarController = () => ctx.get();
