import { mount, onMount, unmount } from "svelte";
import { Timeline, type IdType, type TimelineOptions } from "vis-timeline/esnext";
import { DataSet } from "vis-data/esnext";

import { createQuery, useQueryClient, type CreateQueryResult } from "@tanstack/svelte-query";
import { watch } from "runed";
import { incidentCtx } from "$features/incidents/lib/context.ts";
import {
	listIncidentEventsOptions,
	listIncidentMilestonesOptions,
	type IncidentEvent,
	type ListIncidentEventsResponseBody,
	type ListIncidentMilestonesResponseBody,
} from "$lib/api";
import IncidentTimelineEvent, { type TimelineEventComponentProps } from "./IncidentTimelineEvent.svelte";

const createTimelineEventElement = (event: IncidentEvent) => {
	let props = $state<TimelineEventComponentProps>({ event });

	const target = document.createElement("div");
	target.setAttribute("event-id", $state.snapshot(event.id));

	const component = mount(IncidentTimelineEvent, { target, props });

	return {
		get element() {
			return target;
		},
		unmount: () => (unmount(component)),
	};
};

const createTimelineState = () => {
	let containerEl = $state<HTMLElement>();
	let timeline = $state<Timeline>();

	let milestoneItems = new DataSet<any>([]);
	const eventComponents = new Map<IdType, ReturnType<typeof createTimelineEventElement>>();
	const items = new DataSet<any>([]);

	const clearEventComponents = () => {
		eventComponents.forEach(c => c.unmount());
		eventComponents.clear();
	}

	const onMilestonesQueryDataUpdated = (body?: ListIncidentMilestonesResponseBody) => {
		
	};

	const onEventsQueryDataUpdated = (body?: ListIncidentEventsResponseBody) => {
		if (!body) return;

	};

	const createQueries = () => {
		const queryClient = useQueryClient();
		const incidentId = incidentCtx.get().id;

		const milestonesQueryOptsFn = () => listIncidentMilestonesOptions({ path: { id: incidentId } });
		const milestonesQuery = createQuery(milestonesQueryOptsFn, queryClient);
		watch(() => milestonesQuery.data, onMilestonesQueryDataUpdated);

		const eventsQueryOpts = () => listIncidentEventsOptions({ path: { id: incidentId } });
		const eventsQuery = createQuery(eventsQueryOpts, queryClient);
		watch(() => eventsQuery.data, onEventsQueryDataUpdated);
	};

	// const addEvent = (id: IdType) => {
	// 	const created = createTimelineEventElement(id.toString());
	// 	items.add({ id: 1, content: created.element, start: new Date(2025, 1, 12, 7) });
	// 	eventComponents.set(id, created);
	// };

	const mountContainer = (el?: HTMLElement) => {
		if (!el) return;
		containerEl = el;
		const options: TimelineOptions = {
			height: "100%",
		};
		timeline = new Timeline(containerEl, items, options);
	};

	const onUnmount = () => {
		timeline?.destroy();
		clearEventComponents();
		items.clear();
	};

	const setup = (containerElFn: () => HTMLElement | undefined) => {
		createQueries();
		watch(containerElFn, mountContainer);
		onMount(() => onUnmount);
	};

	let editingEvent = $state<IncidentEvent>();

	return {
		setup,
		get editingEvent() {
			return editingEvent;
		},
	};
};
export const timeline = createTimelineState();
