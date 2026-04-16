import type { RouteId } from "$app/types";

export type SidebarItem = {
    label: string;
    icon?: string;
    route: RouteId;
    subItems?: {
        label: string;
        icon?: string;
        route: RouteId;
    }[];
};
export type SidebarGroup = {
    label?: string;
    items: SidebarItem[];
};
export type SidebarItemActiveStatus = {
    active: boolean;
    subItemsActive: Map<number, true | undefined>;
}
const isActive = (pageRoute: RouteId | null, itemRoute: RouteId) => {
    if (!pageRoute) return false;
    if (itemRoute === "/" || pageRoute === "/" || pageRoute === "/(index)") {
        return itemRoute == pageRoute || (itemRoute === "/" && pageRoute === "/(index)");
    }
    return pageRoute.startsWith(itemRoute);
}
export const getActiveStatus = (pageRoute: RouteId | null, item: SidebarItem) => {
    return {
        active: isActive(pageRoute, item.route),
        subItemsActive: new Map(item.subItems?.map((si, idx) => [idx, isActive(pageRoute, si.route) ? true : undefined])),
    } as SidebarItemActiveStatus;
}