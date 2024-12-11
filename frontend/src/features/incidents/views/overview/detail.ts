import type { Incident } from "$lib/api";

export interface IncidentDetailProps {
	incident: Incident;
	invalidateQuery: () => void;
};