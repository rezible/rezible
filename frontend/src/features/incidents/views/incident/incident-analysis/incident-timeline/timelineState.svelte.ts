import { hydrate, mount, onMount, tick, unmount } from "svelte";
import { Timeline, type DataGroup, type DataItemCollectionType, type TimelineItem, type TimelineOptions } from "vis-timeline/esnext";
import { DataSet } from "vis-data/esnext";

import { createQuery, QueryClient, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

import { listIncidentEventsOptions, listIncidentMilestonesOptions, type Incident, type IncidentEvent, type IncidentMilestone } from "$lib/api";
import { useIncidentViewState } from "../../viewState.svelte";

import IncidentTimelineEventItemContent, { type Props as TimelineEventComponentProps } from "./IncidentTimelineEventItemContent.svelte";
import IncidentTimelineMilestoneItemContent, { type Props as TimelineMilestoneComponentProps } from "./IncidentTimelineMilestoneItemContent.svelte";
import { SvelteSet } from "svelte/reactivity";

const EventsGroup = "events";
const MilestonesGroup = "milestones";

export type TimelineRange = {
	start: number;
	end: number;
}

const OneHour = 1000 * 60 * 60;

const flushItemsAndRedrawTimeline = (i: DataSet<TimelineItem>, tl?: Timeline) => {
	tick().then(() => {i.flush?.(); tl?.redraw()});
}

class TimelineEventElement {
	props = $state<TimelineEventComponentProps>({ selected: false });
	ref = document.createElement("div");
	component: ReturnType<typeof mount> | undefined;

	constructor(event: IncidentEvent) {
		this.ref.setAttribute("event-id", $state.snapshot(event.id));
		this.props.event = event;
		tick().then(() => {
			this.component = mount(IncidentTimelineEventItemContent, { target: this.ref, props: this.props });	
		})
	}

	unmount() {
		if (this.component) unmount(this.component);
		this.ref.remove();
	}
}

const createTimelineEventItem = (e: IncidentEvent, ref: HTMLElement): TimelineItem => {
	return { 
		id: e.id, 
		start: new Date(e.attributes.timestamp),
		type: "box",
		group: EventsGroup,
		// subgroup: "foo",
		content: ref,
	}
}

export const isEventItem = (item: TimelineItem) => {
	return (item.group === EventsGroup);
}

const getMilestoneIds = (id: string) => {
	return [id + "_bg", id + "_box"]
};

const getBackgroundColorForMilestoneKind = (kind: IncidentMilestone["attributes"]["kind"]) => {
	switch (kind) {
		case "impact": return "background-color: #f6ad55;";
		case "detection": return "background-color:rgb(74, 163, 144);";
		case "investigation": return "background-color:rgb(107, 39, 149);";
		case "mitigation": return "background-color:rgb(136, 186, 61);";
		case "resolution": return "background-color: #48bb78;";
	}
}

const createMilestoneTimelineItems = (el: TimelineMilestoneElement, ms: IncidentMilestone, endTime: Date) => {
	const start = new Date(ms.attributes.timestamp);

	const bgStyles = "opacity: 0.15;" + getBackgroundColorForMilestoneKind(ms.attributes.kind);

	const [bgId, boxId] = getMilestoneIds(ms.id);

	const bgItem: TimelineItem = {
		id: bgId,
		type: "background",
		content: "",
		style: bgStyles,
		start,
		end: endTime,
	};

	const boxItem: TimelineItem = {
		id: boxId,
		type: "point",
		group: MilestonesGroup,
		title: ms.attributes.kind,
		content: el.ref,
		align: "left",
		start,
	}

	return [bgItem, boxItem];
}

const createIncidentWindowTimelineItems = (r: TimelineRange) => {
	const windowKey = "incident-window";
	const windowBg: TimelineItem = {
		id: `${windowKey}-bg`,
		start: r.start, 
		end: r.end,
		type: "background",
		content: "",
		style: "background-color:rgba(74, 163, 144, 0.1);",
	};
	const windowStartPoint: TimelineItem = {
		id: `${windowKey}-start`,
		type: "point",
		title: "",
		content: "Incident Opened",
		align: "left",
		selectable: false,
		start: r.start,
	};
	const windowEndPoint: TimelineItem = {
		id: `${windowKey}-end`,
		type: "point",
		title: "",
		content: "Incident Closed",
		align: "left",
		selectable: false,
		start: r.end,
	};
	return [windowBg, windowStartPoint, windowEndPoint];
}

export const isMilestoneItem = (item: TimelineItem) => {
	return (item.group === MilestonesGroup);
}

class TimelineEventsState {
	items: DataSet<TimelineItem>;
	incidentId = $state.raw("");
	timeline = $state.raw<Timeline>();
	timelineElements = new Map<string, TimelineEventElement>();

	queryClient = $state<QueryClient>();
	eventsQuery = createQuery(() => ({
		...listIncidentEventsOptions({ path: { id: this.incidentId } }),
		enabled: !!this.incidentId,
	}));
	events = $derived(this.eventsQuery.data?.data || []);

	constructor(items: DataSet<TimelineItem>) {
		this.items = items;
		this.queryClient = useQueryClient();
		watch(() => this.events, evs => {this.onEventsDataUpdated(evs)});
	}

	setIncident(inc: Incident) {
		this.incidentId = inc.id;
	}

	setTimeline(t: Timeline) {
		this.timeline = t;
	}

	clear() {
		this.timelineElements.forEach(c => {c.unmount()});
		this.timelineElements.clear();
	}

	onEventsDataUpdated(events: IncidentEvent[]) {
		const eventsMap = new Map(events.map(ev => [ev.id, ev]));
		const removeIds: string[] = [];

		this.items.forEach((item, rawId) => {
			if (!isEventItem(item)) return;
			const id = String(rawId);
			if (!eventsMap.has(id)) removeIds.push(id);
		});

		this.removeEventIds(removeIds)
		eventsMap.forEach(ev => {this.updateEvent(ev)});

		flushItemsAndRedrawTimeline(this.items, this.timeline);
	}

	updateEvent(event: IncidentEvent) {
		let el = this.timelineElements.get(event.id);
		if (!el) {
			el = new TimelineEventElement(event);
			this.timelineElements.set(event.id, el);
		}
		const item = createTimelineEventItem(event, el.ref);
		this.items.update(item);
	}

	removeEventIds(ids: string[]) {
		this.items.remove(ids);
		ids.forEach(id => {
			const el = this.timelineElements.get(id);
			if (el) el.unmount();
			this.timelineElements.delete(id);
		})
	}

	onEventAdded(event: IncidentEvent) {
		this.eventsQuery.refetch();
	}

	setSelected(id: string, selected: boolean) {
		const el = this.timelineElements.get(id);
		if (el) el.props.selected = selected;
	}
}

class TimelineMilestoneElement {
	props = $state<TimelineMilestoneComponentProps>({ selected: false });
	ref = document.createElement("div");
	component: ReturnType<typeof mount> | undefined;

	constructor(milestone: IncidentMilestone) {
		this.ref.setAttribute("milestone-id", $state.snapshot(milestone.id));
		this.props.milestone = milestone;
		tick().then(() => {
			this.component = mount(IncidentTimelineMilestoneItemContent, { target: this.ref, props: this.props });	
		})
	}

	unmount() {
		if (this.component) unmount(this.component);
		this.ref.remove();
	}
}

class TimelineMilestonesState {
	items: DataSet<TimelineItem>;
	timeline = $state.raw<Timeline>();
	timelineElements = new Map<string, TimelineMilestoneElement>();
	incident = $state.raw<Incident>();
	incidentEnd = $derived(this.incident ? new Date(this.incident.attributes.closedAt) : new Date());

	milestonesQuery = createQuery(() => ({
		...listIncidentMilestonesOptions({ path: { id: this.incident?.id ?? "" } }),
		enabled: !!this.incident,
	}));
	milestones = $derived(this.milestonesQuery.data?.data || []);

	constructor(items: DataSet<TimelineItem>) {
		this.items = items;
		watch(() => this.milestones, () => {this.onMilestonesDataUpdated()});
	}

	clear() {
		this.timelineElements.forEach(c => c.unmount());
		this.timelineElements.clear();
	}

	setTimeline(t: Timeline) {this.timeline = t}

	setIncident(inc: Incident) {
		this.incident = inc;
	}

	onMilestonesDataUpdated() {
		const dataMap = new Map(this.milestones.map(m => [m.id, m]));
		const removeIds: string[] = [];

		this.items.forEach((item, rawId) => {
			if (!isMilestoneItem(item)) return;
			const msId = String(rawId).split("_")[0];
			if (!dataMap.get(msId)) removeIds.push(msId);
		});

		this.removeMilestoneIds(removeIds)

		const sortedMilestones = this.milestones.toSorted((a: IncidentMilestone, b: IncidentMilestone) => {
			return new Date(a.attributes.timestamp).valueOf() - new Date(b.attributes.timestamp).valueOf();
		});
		sortedMilestones.forEach((ms, idx, arr) => {
			this.setMilestone(ms, arr.at(idx + 1));
		});

		flushItemsAndRedrawTimeline(this.items, this.timeline);
	};

	setMilestone(ms: IncidentMilestone, nextMs?: IncidentMilestone) {
		let el = this.timelineElements.get(ms.id);
		if (!el) {
			el = new TimelineMilestoneElement(ms);
			this.timelineElements.set(ms.id, el);
		}

		const endDate = nextMs ? new Date(nextMs.attributes.timestamp) : this.incidentEnd;
		const msItems = createMilestoneTimelineItems(el, ms, endDate);
		this.items.update(msItems);
	}

	setSelected(id: string, selected: boolean) {
		const msId = id.split("_")[0];
		const el = this.timelineElements.get(msId);
		if (el) el.props.selected = selected;
	}

	removeMilestoneIds(ids: string[]) {
		let itemIds: string[] = [];
		ids.forEach(id => {
			itemIds.push(...getMilestoneIds(id));
			const el = this.timelineElements.get(id);
			el?.unmount();
			this.timelineElements.delete(id);
		});
		this.items.remove(itemIds);
	}
}

export class TimelineState {
	viewState = useIncidentViewState();
	incident = $derived(this.viewState.incident);

	items = new DataSet<TimelineItem>([]);
	events = new TimelineEventsState(this.items);
	milestones = new TimelineMilestonesState(this.items);

	timeline = $state.raw<Timeline>();

	incidentWindow = $state.raw<TimelineRange>({start: 0, end: 0});
	viewWindow = $state.raw<TimelineRange>({start: 0, end: 0});
	viewBounds = $state.raw<TimelineRange>({start: 0, end: 0});

	selectedItems = new SvelteSet<string>();

	constructor() {
		this.items.clear();

		this.items.on("*", () => {this.onItemsUpdate()});

		watch(() => this.incident, inc => {this.onIncidentUpdate(inc)});

		onMount(() => {return () => {this.cleanup()}});
	}

	mountTimeline(ref: HTMLElement) {
		if (this.timeline) {
			console.log("mounting timeline - destroy existing");
			this.timeline.destroy();
		};

		this.timeline = this.createTimeline(ref, this.items as DataItemCollectionType);

		this.timeline.on("select", e => {this.onTimelineSelect(e)});
		this.timeline.on("rangechange", e => {this.onTimelineRangeChanged(e)});

		this.events.setTimeline(this.timeline);
		this.milestones.setTimeline(this.timeline);

		if (this.incident) {
			this.setIncidentWindow(this.incident);
			this.updateTimelineViewBounds(this.timeline);
		}
	}

	createTimeline(ref: HTMLElement, items: DataItemCollectionType) {
		const timelineOpts: TimelineOptions = {
			height: "100%",
			zoomMin: 1000 * 60 * 60,
			zoomMax: 1000 * 60 * 60 * 24 * 7,
			showCurrentTime: false,
		};

		const timelineGroups: DataGroup[] = [
			{ id: "default", title: "Default", content: "" },
			{ id: EventsGroup, title: "Events", content: "" },
			{ id: MilestonesGroup, title: "Milestones", content: "" },
		];

		return new Timeline(ref, items, timelineGroups, timelineOpts);
	}

	onItemsUpdate() {
		if (!this.timeline) return;
		this.updateTimelineViewBounds(this.timeline);
	}

	onIncidentUpdate(inc?: Incident) {
		if (!inc) return;
		this.events.setIncident(inc);
		this.milestones.setIncident(inc);
		this.setIncidentWindow(inc);
	}

	setIncidentWindow(inc: Incident) {
		if (!this.timeline) return;

		const start = new Date(inc.attributes.openedAt).valueOf();
		const end = new Date(inc.attributes.closedAt).valueOf();

		this.incidentWindow = {start, end};
		this.items.update(createIncidentWindowTimelineItems(this.incidentWindow));
		this.timeline.setWindow(start - OneHour, end + OneHour, {animation: false});
	}

	updateTimelineViewBounds(tl: Timeline) {
		let min = this.incidentWindow.start.valueOf();
		let max = this.incidentWindow.end.valueOf();
		this.items.forEach(item => {
			const d = new Date(item.start.valueOf()).valueOf();
			if (d < min) min = d;
			if (d > max) max = d;
		});

		// TODO: offset as % of range?
		const offset = 2 * OneHour;

		const offsetMin = min - offset;
		const offsetMax = max + offset;

		tl.setOptions({min: offsetMin, max: offsetMax});
		this.viewBounds = {start: offsetMin, end: offsetMax};
	}

	cleanup() {
		this.items.clear();
		this.events.clear();
		this.milestones.clear();
		this.timeline?.destroy();
	}

	onTimelineSelect(e: any) {
		const newSelected = new Set(e.items as string[]);
		const deselectedItems = this.selectedItems.difference(newSelected);
		
		deselectedItems.forEach(id => {
			this.events.setSelected(id, false);
			this.milestones.setSelected(id, false);
			this.selectedItems.delete(id);
		});
		newSelected.forEach(id => {
			this.events.setSelected(id, true);
			this.milestones.setSelected(id, true);
			this.selectedItems.add(id);
		});
	}

	onTimelineRangeChanged(e: any) {
		const start = (e.start as Date).valueOf();
		const end = (e.end as Date).valueOf();
		this.viewWindow = {start, end};
	}

	onEventAdded(event: IncidentEvent) {
		this.events.onEventAdded(event);
	}

	onMilestoneAdded(ms: IncidentMilestone) {
		// this.milestones.onMilestoneAdded(ms);
	}
};

const timelineCtx = new Context<TimelineState>("timelineCtx");
export const setIncidentTimeline = (s: TimelineState) => timelineCtx.set(s);
export const useIncidentTimeline = () => timelineCtx.get();