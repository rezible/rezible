import { Context } from "runed";
import type { Component } from "svelte";

import { useSidebar } from "$components/ui/sidebar";

import RiFire from "remixicon-svelte/icons/fire-fill";
import RiHome from "remixicon-svelte/icons/home-2-fill";
import RiSettings3Line from "remixicon-svelte/icons/settings-3-line";
import { useUserSessionState } from "$src/lib/user-session.svelte";
import { useAppShell } from "$src/lib/app-shell.svelte";
import { page } from "$app/state";

export type SidebarItem = {
    label: string;
    icon?: Component;
    href: string;
    subItems?: SidebarItem[];
};

export type SidebarGroup = {
    label?: string;
    items: SidebarItem[];
};

export type SidebarSearch = {
    placeholder: string;
};

export type SidebarModel = {
    isDefault?: boolean;
    search?: SidebarSearch;
    groups: SidebarGroup[];
    footerItems?: SidebarItem[];
};

const isActive = (href: string, pathname: string) => {
    if (href === "/" || pathname === "/") return pathname === href;
    return pathname === href || (pathname.startsWith(`${href}/`));
};

export const makeActiveStatus = (groups: SidebarGroup[], pathname: string) => {
    let deepestActiveItem = "";
    let activeSubItems = new Map<string, Set<string>>();
    groups.forEach(group => {
       group.items.forEach(item => {
            if (item.href.length > deepestActiveItem.length && isActive(item.href, pathname)) {
                deepestActiveItem = item.href;
            };
            const subActive = new Set<string>();
            item.subItems?.forEach(sub => {
                if (isActive(sub.href, pathname)) subActive.add(sub.href);
            });
            activeSubItems.set(item.href, subActive);
       });
    });
    return { deepestActiveItem, activeSubItems };
};

const filterGroups = (groups: SidebarGroup[] | undefined, query?: string) => {
    if (!groups) return [];
    if (!query) return groups;
    return groups
        .map((group) => {
            const labelMatches = group.label?.toLowerCase().includes(query);
            const items = group.items.filter(
                (item) =>
                    labelMatches ||
                    item.label.toLowerCase().includes(query) ||
                    item.subItems?.some((subItem) =>
                        subItem.label.toLowerCase().includes(query),
                    ),
            );
            return { ...group, items };
        })
        .filter((group) => group.items.length > 0);
};

export const defaultSidebarModel: SidebarModel = {
    groups: [
        {
            items: [
                { label: "Home", href: "/", icon: RiHome }
            ]
        },
        {
            label: "System",
            items: [
                { label: "Incidents", href: "/incidents", icon: RiFire },
            ],
        },
    ],
    footerItems: [
        { label: "Settings", href: "/settings", icon: RiSettings3Line },
    ],
};

export class AppSidebarController {
    private shell = useAppShell();
    private sidebarState = useSidebar();
    private session = useUserSessionState();

    isOpen = $derived(this.sidebarState.open);
    collapsed = $derived(this.sidebarState.state === "collapsed");

    isDefault = $derived(!this.shell.childSidebar);
    model = $derived(!this.shell.childSidebar ? defaultSidebarModel : this.shell.childSidebar);

    searchQuery = $state("");

	private normalizedQuery = $derived(this.searchQuery.trim().toLowerCase());

    showSearch = $derived(!!this.model.search);
	groups = $derived(this.showSearch ? filterGroups(this.model.groups, this.normalizedQuery) : this.model.groups);
    preloadHome = $derived<"tap" | "hover">(this.session.error ? "tap" : "hover");

    activeStatus = $derived(makeActiveStatus(this.groups, page.url.pathname)); 
}

const ctx = new Context<AppSidebarController>("AppSidebarController");
export const initAppSidebarController = () => ctx.set(new AppSidebarController());
export const useAppSidebarController = () => ctx.get();
