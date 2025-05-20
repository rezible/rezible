import { hydrate, mount, onMount, tick, unmount } from "svelte";
import { Timeline, type DataGroup, type DataItemCollectionType, type TimelineItem, type TimelineOptions } from "vis-timeline/esnext";
import { DataSet } from "vis-data/esnext";

import { createQuery, QueryClient, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

import { listIncidentEventsOptions, listIncidentMilestonesOptions, type Incident, type IncidentEvent, type IncidentMilestone } from "$lib/api";
import { useIncidentViewState } from "../../viewState.svelte";

import IncidentTimelineEventItemContent, { type Props as TimelineEventComponentProps } from "./IncidentTimelineEventItemContent.svelte";
import IncidentTimelineMilestoneItemContent, { type Props as TimelineMilestoneComponentProps } from "./IncidentTimelineMilestoneItemContent.svelte";
import { SvelteMap, SvelteSet } from "svelte/reactivity";

const EventsSubgroup = "events";
const MilestonesSubgroup = "milestones";

const flushItemsAndRedrawTimeline = (i: DataSet<TimelineItem>, tl?: Timeline) => {
	tick().then(() => {
		if (i.flush) i.flush();
		if (tl) tl.redraw()
	});
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

class TimelineEventsState {
	items: DataSet<TimelineItem>;
	incidentId = $state("");
	timeline = $state<Timeline>();
	timelineElements = new Map<string, TimelineEventElement>();

	queryClient = $state<QueryClient>();
	eventsQuery = createQuery(() => ({
		...listIncidentEventsOptions({ path: { id: this.incidentId } }),
		enabled: !!this.incidentId,
	}));
	events = $derived(this.eventsQuery.data?.data ?? []);

	constructor(items: DataSet<TimelineItem>) {
		this.items = items;
		this.queryClient = useQueryClient();
		watch(() => this.events, evs => {this.onEventsDataUpdated(evs)});
	}

	setIncidentId(id: string) {
		this.incidentId = id;
	}

	setTimeline(t: Timeline) {
		this.timeline = t;
	}

	clear() {
		this.timelineElements.forEach(c => {c.unmount()});
		this.timelineElements.clear();
	}

	updateEvent(event: IncidentEvent) {
		let el = this.timelineElements.get(event.id);
		if (!el) {
			el = new TimelineEventElement(event);
			this.timelineElements.set(event.id, el);
		}
		this.items.update({ 
			id: event.id, 
			start: new Date(event.attributes.timestamp),
			type: "box",
			group: "default",
			subgroup: EventsSubgroup,
			content: el.ref,
		});
	}

	removeEvent(id: string) {
		const el = this.timelineElements.get(id);
		if (el) el.unmount();
		this.items.remove(id);
		this.timelineElements.delete(id);
	}

	onEventsDataUpdated(events: IncidentEvent[]) {
		const wasEmpty = this.timelineElements.size === 0;
		const eventsMap = new Map(events.map(ev => [ev.id, ev]));
		const removeIds: string[] = [];

		this.items.forEach((item, rawId) => {
			if (item.subgroup !== "events") return;
			const id = String(rawId);
			if (eventsMap.has(id)) return;
			removeIds.push(id);
		});

		removeIds.forEach(id => {this.removeEvent(id)});
		eventsMap.forEach(ev => {this.updateEvent(ev)});

		if (wasEmpty) flushItemsAndRedrawTimeline(this.items, this.timeline);
	}

	eventAdded(event: IncidentEvent) {
		console.log("event added");
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
	milestones = $derived(this.milestonesQuery.data?.data ?? []);

	constructor(items: DataSet<TimelineItem>) {
		this.items = items;
		watch(() => this.milestones, () => this.onMilestonesDataUpdated());
	}

	clear() {
		this.timelineElements.forEach(c => c.unmount());
		this.timelineElements.clear();
	}

	setTimeline(t: Timeline) {this.timeline = t}

	setIncident(inc: Incident) {this.incident = inc}

	onMilestonesDataUpdated() {
		const wasEmpty = this.timelineElements.size === 0;
		const dataMap = new Map(this.milestones.map(m => [m.id, m]));
		const removeIds: string[] = [];

		this.items?.forEach((item, rawId) => {
			if (item.subgroup !== "milestones") return;
			const msId = String(rawId).split("_")[0];
			if (!dataMap.get(msId)) removeIds.push(msId);
		});

		removeIds.forEach(id => {this.removeMilestone(id)});

		const sortedMilestones = this.milestones.toSorted((a: IncidentMilestone, b: IncidentMilestone) => {
			return new Date(a.attributes.timestamp).valueOf() - new Date(b.attributes.timestamp).valueOf();
		});
		sortedMilestones.forEach((ms, idx, arr) => {
			this.setMilestone(ms, idx < arr.length - 1 ? arr[idx + 1] : undefined);
		});

		if (wasEmpty) flushItemsAndRedrawTimeline(this.items, this.timeline);
	};

	getMilestoneIds(id: string) {
		return [id + "_bg", id + "_box"]
	};

	makeMilestoneTimelineItems(el: TimelineMilestoneElement, ms: IncidentMilestone, endTime: Date) {
		const start = new Date(ms.attributes.timestamp);
	
		const getBackgroundColorForMilestoneKind = (kind: IncidentMilestone["attributes"]["kind"]) => {
			switch (kind) {
				case "impact": return "background-color: #f6ad55;";
				case "detection": return "background-color:rgb(74, 163, 144);";
				case "investigation": return "background-color:rgb(107, 39, 149);";
				case "mitigation": return "background-color:rgb(136, 186, 61);";
				case "resolution": return "background-color: #48bb78;";
			}
		}
	
		const bgStyles = "opacity: 0.15;" + getBackgroundColorForMilestoneKind(ms.attributes.kind);
	
		const [bgId, boxId] = this.getMilestoneIds(ms.id);
	
		const bgItem: TimelineItem = {
			id: bgId,
			type: "background",
			content: "foo",
			style: bgStyles,
			start,
			end: endTime,
		};
	
		const boxItem: TimelineItem = {
			id: boxId,
			type: "point",
			group: "default",
			subgroup: MilestonesSubgroup,
			title: ms.attributes.kind,
			content: el.ref,
			align: "left",
			start,
		}
	
		return [bgItem, boxItem];
	}

	setMilestone(ms: IncidentMilestone, nextMs?: IncidentMilestone) {
		let el = this.timelineElements.get(ms.id);
		if (!el) {
			el = new TimelineMilestoneElement(ms);
			this.timelineElements.set(ms.id, el);
		}

		const endDate = nextMs ? new Date(nextMs.attributes.timestamp) : this.incidentEnd;
		const msItems = this.makeMilestoneTimelineItems(el, ms, endDate);
		this.items.update(msItems);
	}

	setSelected(id: string, selected: boolean) {
		const msId = id.split("_")[0];
		const el = this.timelineElements.get(msId);
		if (el) el.props.selected = selected;
	}

	removeMilestone(id: string) {
		const el = this.timelineElements.get(id);
		if (el) el.unmount();
		this.items.remove(this.getMilestoneIds(id));
		this.timelineElements.delete(id);
	}
}

export type TimelineRange = {
	start: Date;
	end: Date;
}

export class TimelineState {
	viewState = useIncidentViewState();

	items = new DataSet<TimelineItem>([]);
	events = new TimelineEventsState(this.items);
	milestones = new TimelineMilestonesState(this.items);

	timeline = $state.raw<Timeline>();

	incidentWindow = $state.raw<TimelineRange>();
	viewWindow = $state.raw<TimelineRange>();

	selectedItems = new SvelteSet<string>();

	constructor() {
		this.items.clear();

		watch(() => this.viewState.incident, inc => {
			if (!inc) return;

			this.events.setIncidentId(inc.id);
			this.milestones.setIncident(inc);
			
			this.setIncidentWindow($state.snapshot(inc));
		});

		onMount(() => {
			return () => {this.cleanup()}
		});
	}

	mountTimeline(ref: HTMLElement) {
		const timelineOpts: TimelineOptions = {
			height: "100%",
			zoomMin: 1000 * 60 * 60,
			zoomMax: 1000 * 60 * 60 * 24 * 7,
		};

		const timelineGroups: DataGroup[] = [
			{ id: "default", title: "Default", content: "" },
		];

		if (this.timeline) this.timeline.destroy();

		this.timeline = new Timeline(ref, this.items as DataItemCollectionType, timelineGroups, timelineOpts);
		this.events.setTimeline(this.timeline);
		this.milestones.setTimeline(this.timeline);

		this.timeline.on("select", e => {this.onTimelineSelect(e)});
		this.timeline.on("rangechange", e => {
			this.viewWindow = {start: e.start as Date, end: e.end as Date};
		});

		this.setIncidentWindow(this.viewState.incident);
	}

	setIncidentWindow(inc?: Incident) {
		if (!inc || !this.timeline) return;

		const start = new Date(inc.attributes.openedAt);
		const end = new Date(inc.attributes.closedAt);

		this.incidentWindow = {start, end};

		const hour = 1000 * 60 * 60;		
		this.timeline.setWindow(start.valueOf() - hour, end.valueOf() + hour, {animation: false});

		const windowKey = "incident-window-bg";
		const windowBackgroundStyle = "background-color:rgba(74, 163, 144, 0.5);";
		this.items.update({
			id: windowKey,
			start, end,
			type: "background",
			style: windowBackgroundStyle,
		});

		this.items.update({
			id: "incident-window-start",
			type: "point",
			title: "Incident Opened",
			content: "Incident Opened",
			align: "left",
			selectable: false,
			start,
		});
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

	onEventAdded(event: IncidentEvent) {
		this.events.eventAdded(event);
	}
};

const timelineCtx = new Context<TimelineState>("timelineCtx");
export const setIncidentTimeline = (s: TimelineState) => timelineCtx.set(s);
export const useIncidentTimeline = () => timelineCtx.get();