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
	if (i.flush) i.flush();
	if (tl) tick().then(() => {if (tl) tl.redraw()});
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
	timelineElements = new SvelteMap<string, TimelineEventElement>();

	queryClient = $state<QueryClient>();
	
	constructor(items: DataSet<TimelineItem>) {
		this.items = items;
		this.clearTimelineElements();
		this.queryClient = useQueryClient();
		watch(() => this.events, e => {this.onEventsDataUpdated()});
	}

	setIncidentId(id: string) {
		this.incidentId = id;
	}

	setTimeline(t: Timeline) {
		this.timeline = t;
	}

	eventsQuery = createQuery(() => ({
		...listIncidentEventsOptions({ path: { id: this.incidentId } }),
		enabled: !!this.incidentId,
	}));
	events = $derived(this.eventsQuery.data?.data ?? []);

	clearTimelineElements() {
		this.timelineElements.forEach(c => c.unmount());
		this.timelineElements.clear();
	}

	setEvent(event: IncidentEvent) {
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

	onEventsDataUpdated() {
		const eventsMap = new Map(this.events.map(ev => [ev.id, ev]));
		const removeIds: string[] = [];

		this.items.forEach((item, rawId) => {
			if (item.subgroup !== "events") return;
			const id = String(rawId);
			if (eventsMap.has(id)) return;
			removeIds.push(id);
		});

		removeIds.forEach(id => {this.removeEvent(id)});
		eventsMap.forEach(ev => {this.setEvent(ev)});

		flushItemsAndRedrawTimeline(this.items, this.timeline);
	}

	eventAdded(event: IncidentEvent) {
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
	timeline = $state<Timeline>();

	timelineElements = new Map<string, TimelineMilestoneElement>();

	constructor(items: DataSet<TimelineItem>) {
		this.items = items;
		this.clearTimelineElements();
		watch(() => this.milestones, () => this.onMilestonesDataUpdated());
	}

	clearTimelineElements() {
		this.timelineElements.forEach(c => c.unmount());
		this.timelineElements.clear();
	}

	incident = $state<Incident>();
	incidentEnd = $derived(this.incident ? new Date(this.incident.attributes.closedAt) : new Date());

	setIncident(inc: Incident) {this.incident = inc}

	setTimeline(t: Timeline) {this.timeline = t}

	milestonesQuery = createQuery(() => ({
		...listIncidentMilestonesOptions({ path: { id: this.incident?.id ?? "" } }),
		enabled: !!this.incident,
	}));
	milestones = $derived(this.milestonesQuery.data?.data ?? []);

	onMilestonesDataUpdated() {
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

		flushItemsAndRedrawTimeline(this.items, this.timeline);
	};

	getMilestoneIds(id: string) {
		return {bg: id + "_bg", box: id + "_box"}
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
	
		const ids = this.getMilestoneIds(ms.id);
	
		const bgItem: TimelineItem = {
			id: ids.bg,
			type: "background",
			content: "",
			style: bgStyles,
			start,
			end: endTime,
		};
	
		const boxItem: TimelineItem = {
			id: ids.box,
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
		// if (!!this.items.get(msItems[0].id)) {
		// 	this.items.update(msItems);
		// } else {
		// 	this.items.add(msItems);
		// }
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
		const ids = this.getMilestoneIds(id);
		this.items.remove([ids.bg, ids.box]);
		this.timelineElements.delete(id);
	}
}

export class TimelineState {
	viewState = useIncidentViewState();

	items = new DataSet<TimelineItem>([]);
	events = new TimelineEventsState(this.items);
	milestones = new TimelineMilestonesState(this.items);

	timeline = $state.raw<Timeline>();
	selectedItems = new SvelteSet<string>();

	setIncidentWindow(inc?: Incident) {
		if (!inc || !this.timeline) return;

		const openedAt = new Date(inc.attributes.openedAt).valueOf();
		const closedAt = new Date(inc.attributes.closedAt).valueOf();

		const hour = 100 * 60 * 60;		
		this.timeline.setWindow(openedAt - hour, closedAt + hour, {animation: false});
	}

	constructor() {
		this.items.clear();

		watch(() => this.viewState.incident, inc => {
			if (!inc) return;

			this.events.setIncidentId(inc.id);
			this.milestones.setIncident(inc);
			
			this.setIncidentWindow(inc);
		});

		onMount(() => {
			return this.cleanup()
		});
	}

	mountTimeline(ref: HTMLElement) {
		const timelineOpts: TimelineOptions = {
			height: "100%",
			zoomMin: 1000 * 60 * 60,
			zoomMax: 1000 * 60 * 60 * 24,
		};

		const timelineGroups: DataGroup[] = [
			{ id: "default", title: "Incident", content: "", subgroupStack: { "incident": true } },
		];

		if (this.timeline) this.timeline.destroy();

		this.timeline = new Timeline(ref, this.items as DataItemCollectionType, timelineGroups, timelineOpts);
		this.events.setTimeline(this.timeline);
		this.milestones.setTimeline(this.timeline);

		this.timeline.on("select", e => {this.onTimelineSelect(e)});

		this.setIncidentWindow(this.viewState.incident);
	}

	cleanup() {
		this.timeline?.destroy();
		this.events.clearTimelineElements();
		this.milestones.clearTimelineElements();
		this.items.clear();
	}

	onTimelineSelect(e: any) {
		const newSelected = new Set(e.items as string[]);
		const deselectedItems = this.selectedItems.difference(newSelected);
		
		deselectedItems.forEach(id => {
			this.selectedItems.delete(id);
			this.events.setSelected(id, false);
			this.milestones.setSelected(id, false);
		});
		newSelected.forEach(id => {
			this.selectedItems.add(id);
			this.events.setSelected(id, true);
			this.milestones.setSelected(id, true);
		});
	}

	onEventAdded(event: IncidentEvent) {
		this.events.eventAdded(event);
	}
};

const timelineCtx = new Context<TimelineState>("timelineCtx");
export const setIncidentTimeline = (s: TimelineState) => timelineCtx.set(s);
export const useIncidentTimeline = () => timelineCtx.get();