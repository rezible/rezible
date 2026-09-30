import { IsMobile } from "$lib/hooks/is-mobile.svelte.js";
import { getContext, setContext } from "svelte";

type Getter<T> = () => T;

export type SidebarStateProps = {
	/** Desktop only: collapse to the icon rail. Decided by the page, not toggled by the user. */
	collapsed: Getter<boolean>;
};

class SidebarState {
	#isMobile: IsMobile;

	constructor(readonly props: SidebarStateProps) {
		this.#isMobile = new IsMobile();
	}

	// Convenience getter for checking if the sidebar is mobile
	// without this, we would need to use `sidebar.isMobile.current` everywhere
	get isMobile() {
		return this.#isMobile.current;
	}

	openMobile = $state(false);

	/** Whether labels are shown: always in the mobile sheet, and on desktop unless collapsed. */
	get expanded() {
		return this.isMobile || !this.props.collapsed();
	}

	get state() {
		return this.expanded ? "expanded" : "collapsed";
	}

	setOpenMobile = (value: boolean) => {
		this.openMobile = value;
	};

	/** Opens or closes the mobile sheet. The desktop sidebar has no toggle. */
	toggle = () => {
		this.openMobile = !this.openMobile;
	};
}

const SYMBOL_KEY = "scn-sidebar";

/**
 * Instantiates a new `SidebarState` instance and sets it in the context.
 *
 * @param props The constructor props for the `SidebarState` class.
 * @returns  The `SidebarState` instance.
 */
export function setSidebar(props: SidebarStateProps): SidebarState {
	return setContext(Symbol.for(SYMBOL_KEY), new SidebarState(props));
}

/**
 * Retrieves the `SidebarState` instance from the context. This is a class instance,
 * so you cannot destructure it.
 * @returns The `SidebarState` instance.
 */
export function useSidebar(): SidebarState {
	return getContext(Symbol.for(SYMBOL_KEY));
}
