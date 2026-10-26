/** Returns the URL with the path segment `from` swapped for `to`, keeping the view, search and hash. */
export function referenceUrl(url: URL, from: string, to: string) {
	const segments = url.pathname.split("/");
	const index = segments.indexOf(encodeURIComponent(from));
	if (index !== -1) {
		segments[index] = encodeURIComponent(to);
	}
	return `${segments.join("/")}${url.search}${url.hash}`;
}
