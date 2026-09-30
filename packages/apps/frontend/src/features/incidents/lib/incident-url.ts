export function incidentSlugUrl(url: URL, id: string, slug: string) {
	const path = url.pathname.replace(`/incidents/${id}`, `/incidents/${encodeURIComponent(slug)}`);
	return `${path}${url.search}${url.hash}`;
}
