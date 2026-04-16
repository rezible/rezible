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
export type ActiveSidebarItem = {
    itemRoute?: RouteId;
    subItemRoute?: RouteId;
}
export const getActiveSidebarItem = (groups: SidebarGroup[], pageRoute: RouteId | null): ActiveSidebarItem => {
    if (!pageRoute) return {};
    let deepestRoute: RouteId | undefined;
    let deepestSubRoute: RouteId | undefined;
    const isDeeperActive = (r: RouteId, curr?: RouteId) => {
        if (!pageRoute.startsWith(r)) return false;
        return !curr || curr.length < r.length;
    }
    for (let gi = 0; gi < groups.length; gi++) {
        for (let i = 0; i < groups[gi].items.length; i++) {
            const { route, subItems } = groups[gi].items[i];
            if (pageRoute !== "/" && isDeeperActive(route, deepestRoute)) {
                deepestRoute = route;
            }
            if (!subItems) continue;
            for (let si = 0; si < subItems.length; si++) {
                const { route: subRoute } = subItems[si];
                if (isDeeperActive(subRoute, deepestSubRoute)) {
                    deepestRoute = route;
                    deepestSubRoute = subRoute;
                }
            }
        }
    }
    return {
        itemRoute: deepestRoute,
        subItemRoute: deepestSubRoute,
    }
}
export const findActiveSidebarSubItem = (subItems: SidebarItem["subItems"], activeSubRoute?: RouteId) => {
    if (!activeSubRoute || !subItems) return -1;
    return subItems.findIndex(i => i.route === activeSubRoute);
}