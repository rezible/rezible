import { mount, onMount, unmount } from "svelte";
import { Timeline, type DataGroup, type DataItem, type DataItemCollectionType, type IdType, type TimelineItem, type TimelineOptions } from "vis-timeline/esnext";
import { DataSet } from "vis-data/esnext";

import { createQuery, QueryClient, useQueryClient, type CreateQueryResult } from "@tanstack/svelte-query";
import { watch } from "runed";
import { incidentCtx } from "$features/incidents/lib/context.ts";
import {
	listIncidentEventsOptions,
	listIncidentMilestonesOptions,
	type Incident,
	type IncidentEvent,
	type IncidentMilestone,
	type ListIncidentEventsResponseBody,
	type ListIncidentMilestonesResponseBody,
} from "$lib/api";
import IncidentTimelineEvent, { type TimelineEventComponentProps } from "./IncidentTimelineEvent.svelte";
import { SvelteMap } from "svelte/reactivity";

const createTimelineElement = (event: IncidentEvent) => {
	let props = $state<TimelineEventComponentProps>({ event });

	const target = document.createElement("div");
	target.setAttribute("event-id", $state.snapshot(event.id));

	const component = mount(IncidentTimelineEvent, { target, props });

	return {
		get ref() {
			return target;
		},
		get props() { return props },
		set props(newProps: TimelineEventComponentProps) { props = newProps },
		unmount: () => (unmount(component)),
	};
};
type TimelineEventElement = ReturnType<typeof createTimelineElement>;

const createTimelineState = () => {
	let incident = $state<Incident>();
	// const incidentEnd = $derived(incident ? new Date(incident.attributes.closedAt) : new Date(Date.now() + 1000));
	const incidentId = $derived(incident?.id ?? "");
	let queryClient = $state<QueryClient>();

	const items = new DataSet<TimelineItem>([]);
	let timeline = $state<Timeline>();

	const listMilestonesQueryOpts = $derived(listIncidentMilestonesOptions({ path: { id: incidentId } }));
	let milestonesQuery = $state<CreateQueryResult<ListIncidentMilestonesResponseBody>>();
	const milestones = $derived(milestonesQuery?.data?.data ?? []);

	const listEventsQueryOpts = $derived(listIncidentEventsOptions({ path: { id: incidentId } }));
	let eventsQuery = $state<CreateQueryResult<ListIncidentEventsResponseBody>>();
	const events = $derived(eventsQuery?.data?.data ?? []);
	const timelineEventElements = $state(new SvelteMap<string, TimelineEventElement>());

	const makeTimelineEventItem = (el: TimelineEventElement, event: IncidentEvent) => {
		const start = new Date(event.attributes.timestamp);
		return { id: event.id, content: el.ref, start } as TimelineItem;
	}

	const addEvent = (event: IncidentEvent) => {
		const id = event.id;
		const el = createTimelineElement(event);
		timelineEventElements.set(id, el);
		items.add(makeTimelineEventItem(el, event));
	}

	const maybeUpdateEvent = (el: TimelineEventElement, event: IncidentEvent) => {
		if (el.props.event.attributes.timestamp === event.attributes.timestamp) return;
		items.update(makeTimelineEventItem(el, event));
	}

	const removeEvent = (id: string) => {
		const el = timelineEventElements.get(id);
		if (el) el.unmount();
		items.remove(id);
		timelineEventElements.delete(id);
	}

	const clearEventElements = () => {
		timelineEventElements.forEach(c => c.unmount());
		timelineEventElements.clear();
	}

	const onMilestonesDataUpdated = () => {
		const dataMap = new Map(milestones.map(m => [m.id, m]));
		const msExists = new Map(milestones.map(m => [m.id, false]));
		// const removeIds: string[] = [];
		
		items.forEach((item, rawId) => {
			if (item.type !== "background") return;
			const id = String(rawId);
			if (dataMap.get(id)) {
				msExists.set(id, true);
			} else {
				// TODO: maybe don't remove while iterating?
				// removeIds.push(id);
				items.remove(id);
			}
		});

		//removeIds.forEach(id => items.remove(id));

		// TODO: don't constantly make new Date objects
		const ordered = milestones.toSorted((a, b) => new Date(a.attributes.timestamp).valueOf() - new Date(b.attributes.timestamp).valueOf());
		ordered.forEach((milestone, idx, arr) => {
			const start = new Date(milestone.attributes.timestamp);

			const next = idx < arr.length - 1 ? arr[idx + 1] : undefined;
			const end = next ? new Date(next.attributes.timestamp) : new Date(start.valueOf() + 1000000);

			const updated: TimelineItem = { id: milestone.id, type: "background", content: milestone.attributes.title, start, end };
			if (msExists) {
				items.update(updated);
			} else {
				items.add(updated);
			}
		});

		if (items.flush) items.flush();
	};

	const onEventsDataUpdated = () => {
		const eventsMap = new Map(events.map(ev => [ev.id, ev]));
		// const removeIds: string[] = [];

		items.forEach((item, rawId) => {
			if (item.type === "background") return;
			const id = String(rawId);
			if (eventsMap.has(id)) return;
			// TODO: maybe don't remove while iterating?
			// removeIds.push(id);
			removeEvent(id);
		});

		//removeIds.forEach(removeEvent);

		eventsMap.forEach((event, id) => {
			const curr = timelineEventElements.get(id);
			if (curr) {
				maybeUpdateEvent(curr, event);
			} else {
				addEvent(event);
			}
		});

		if (items.flush) items.flush();
	};

	const eventAdded = (event: IncidentEvent) => {
		queryClient?.setQueryData(listEventsQueryOpts.queryKey, current => {
			if (!current) return current;
			return {...current, data: [...current.data, event]};
		});
	};

	const createQueries = () => {
		queryClient = useQueryClient();

		milestonesQuery = createQuery(() => listMilestonesQueryOpts);
		watch(() => milestones, onMilestonesDataUpdated);

		eventsQuery = createQuery(() => listEventsQueryOpts);
		watch(() => events, onEventsDataUpdated);
	};

	const mount = (containerRef: HTMLElement) => {
		const options: TimelineOptions = {
			height: "100%",
		};
		// @ts-expect-error incorrect timeline DataItem typing for content
		timeline = new Timeline(containerRef, items, options);
	};

	const unmount = () => {
		timeline?.destroy();
		clearEventElements();
		items.clear();
	};

	const setup = (containerRefFn: () => HTMLElement | undefined) => {
		incident = incidentCtx.get();

		createQueries();

		watch(containerRefFn, ref => { if (ref) mount(ref) });
		onMount(() => unmount);
	};

	return {
		setup,
		eventAdded,
		get milestones() { return milestones },
		get events() { return events },
	};
};
export const timeline = createTimelineState();
