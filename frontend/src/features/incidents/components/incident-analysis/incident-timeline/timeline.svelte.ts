import { mount, onMount, unmount } from "svelte";
import { Timeline, type IdType, type TimelineOptions } from "vis-timeline/esnext";
import { DataSet } from "vis-data/esnext";

import { createQuery, useQueryClient, type CreateQueryResult } from "@tanstack/svelte-query";
import { watch } from "runed";
import { incidentCtx } from "$features/incidents/lib/context.ts";
import {
	listIncidentMilestonesOptions,
	type IncidentEvent,
	type ListIncidentMilestonesResponseBody,
} from "$lib/api";
import IncidentTimelineEvent, { type TimelineEventComponentProps } from "./IncidentTimelineEvent.svelte";

const createTimelineEventElement = (id: string) => {
	let props = $state<TimelineEventComponentProps>({ label: "example" });

	const target = document.createElement("div");
	target.setAttribute("event-id", id);

	const component = mount(IncidentTimelineEvent, { target, props });

	return {
		get element() {
			return target;
		},
		setLabel: (label: string) => (props.label = label),
		unmount: () => (unmount(component)),
	};
};

const createTimelineState = () => {
	let containerEl = $state<HTMLElement>();
	let timeline = $state<Timeline>();

	let milestoneItems = new DataSet<any>([]);
	const eventComponents = new Map<IdType, ReturnType<typeof createTimelineEventElement>>();
	const items = new DataSet<any>([]);

	const updateItems = () => {
	};

	const onMilestonesQueryDataUpdated = (res: CreateQueryResult<ListIncidentMilestonesResponseBody, Error>) => {
		
	};

	const onEventsQueryDataUpdated = (res: CreateQueryResult<ListIncidentMilestonesResponseBody, Error>) => {
		
	};

	const createQueries = () => {
		const queryClient = useQueryClient();
		const incidentId = incidentCtx.get().id;

		const milestonesQueryOptsFn = () => listIncidentMilestonesOptions({ path: { id: incidentId } });
		const milestonesQuery = createQuery(milestonesQueryOptsFn, queryClient);
		watch(() => milestonesQuery, onMilestonesQueryDataUpdated);

		// TODO: swap this for correct query
		const eventsQueryOpts = () => listIncidentMilestonesOptions({ path: { id: incidentId } });
		const eventsQuery = createQuery(eventsQueryOpts, queryClient);
		watch(() => eventsQuery, onEventsQueryDataUpdated);
	};

	const addEvent = (id: IdType) => {
		const created = createTimelineEventElement(id.toString());
		items.add({ id: 1, content: created.element, start: new Date(2025, 1, 12, 7) });
		eventComponents.set(id, created);
	};

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
		eventComponents.forEach(c => c.unmount());
		eventComponents.clear();
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
