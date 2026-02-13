import { appShell, type PageBreadcrumb as CrumbType } from "./lib/appShellState.svelte";
import AppShell from "./components/app-shell/AppShell.svelte";

export type PageBreadcrumb = CrumbType;
export { AppShell, appShell };