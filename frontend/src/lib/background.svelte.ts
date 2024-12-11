import type { QueryClient } from "@tanstack/svelte-query";

const checkInterval = 5000;

const createAuthSessionCheck = (client: QueryClient) => {
	
};

const createBackgroundTask = (fn: Function, interval = 5000, immediate = true): VoidFunction => {
	if (immediate) fn();
	const i = setInterval(fn, interval)
	return () => {clearInterval(i)}
}

// const backgroundTasks = () => {
// 	const cleanupAuth = createBackgroundTask(() => session.checkRefresh(queryClient));
// 	return () => {
// 		cleanupAuth();
// 	};
// }