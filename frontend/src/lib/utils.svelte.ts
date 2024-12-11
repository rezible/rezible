export function debounce<T extends Function>(cb: T, wait = 100) {
	let timeout: ReturnType<typeof setTimeout>;
	let callable = (...args: any) => {
		clearTimeout(timeout);
		timeout = setTimeout(() => cb(...args), wait);
	};
	return <T>(<any>callable);
}
