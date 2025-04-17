import { mount, onMount, unmount } from "svelte";
import { Timeline, type DataGroup, type DataItemCollectionType, type TimelineItem, type TimelineOptions } from "vis-timeline/esnext";
import { DataSet } from "vis-data/esnext";

import { createQuery, QueryClient, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

import { listIncidentEventsOptions, listIncidentMilestonesOptions, type Incident, type IncidentEvent, type IncidentMilestone } from "$lib/api";
import { useIncidentViewState } from "../../viewState.svelte";

import IncidentTimelineEventItemContent, { type Props as TimelineEventComponentProps } from "./IncidentTimelineEventItemContent.svelte";
import IncidentTimelineMilestoneItemContent, { type Props as TimelineMilestoneComponentProps } from "./IncidentTimelineMilestoneItemContent.svelte";
import { SvelteSet } from "svelte/reactivity";

const EventsSubgroup = "events";
const MilestonesSubgroup = "milestones";

const createTimelineEventElement = (event: IncidentEvent) => {
	let props = $state<TimelineEventComponentProps>({ event, selected: false });

	const target = document.createElement("div");
	target.setAttribute("event-id", $state.snapshot(event.id));

	const component = mount(IncidentTimelineEventItemContent, { target, props });

	return {
		get ref() { return target },
		get props() { return props },
		set props(newProps: TimelineEventComponentProps) { props = newProps },
		unmount: () => (unmount(component)),
	};
};

// const fitTimeline = debounce((t?: Timeline) => {
// 	try {
// 		t?.fit({ zoom: true, animation: { duration: 500, easingFunction: "easeInOutQuad" } })
// 	} catch (e) {}
// }, 100);

type TimelineEventElement = ReturnType<typeof createTimelineEventElement>;
const makeTimelineEventItem = (el: TimelineEventElement, event: IncidentEvent) => {
	const start = new Date(event.attributes.timestamp);
	return { id: event.id, type: "box", group: "default", subgroup: EventsSubgroup, content: el.ref, start } as TimelineItem;
}

class TimelineEventsState {
	items = $state<DataSet<TimelineItem>>();
	incidentId = $state("");
	timeline = $state<Timeline>();
	timelineElements = new Map<string, TimelineEventElement>();

	queryClient = $state<QueryClient>();
	
	constructor(items: DataSet<TimelineItem>) {
		this.items = items;
		this.clearTimelineElements();
		this.queryClient = useQueryClient();
		watch(() => this.events, e => this.onEventsDataUpdated());
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
		const curr = this.timelineElements.get(event.id);
		if (curr /* && needsUpdate */) {
			this.items?.update(makeTimelineEventItem(curr, event));
		} else {
			const el = createTimelineEventElement(event);
			this.timelineElements.set(event.id, el);
			this.items?.add(makeTimelineEventItem(el, event));
		}
	}

	removeEvent(id: string) {
		const el = this.timelineElements.get(id);
		if (el) el.unmount();
		this.items?.remove(id);
		this.timelineElements.delete(id);
	}

	onEventsDataUpdated() {
		const eventsMap = new Map(this.events?.map(ev => [ev.id, ev]));
		const removeIds: string[] = [];

		this.items?.forEach((item, rawId) => {
			if (item.subgroup !== "events") return;
			const id = String(rawId);
			if (eventsMap.has(id)) return;
			removeIds.push(id);
		});

		removeIds.forEach(this.removeEvent);
		eventsMap.forEach(this.setEvent);

		if (this.items?.flush) this.items.flush();
	}

	eventAdded(event: IncidentEvent) {
		// this.queryClient?.setQueryData(listEventsQueryOpts.queryKey, current => {
		// 	if (!current) return current;
		// 	return { ...current, data: [...current.data, event] };
		// });
		// queryClient.invalidateQueries(listEventsQueryOpts);
		this.eventsQuery.refetch();
	}

	setSelected(id: string, selected: boolean) {
		const el = this.timelineElements.get(id);
		if (!el) return;
		el.props.selected = selected;
	}
}

const createTimelineMilestoneElement = (milestone: IncidentMilestone) => {
	let props = $state<TimelineMilestoneComponentProps>({ milestone, selected: false });

	const target = document.createElement("div");
	target.setAttribute("milestone-id", $state.snapshot(milestone.id));

	const component = mount(IncidentTimelineMilestoneItemContent, { target, props });

	return {
		get ref() { return target },
		get selected() { return props.selected },
		set selected(s: boolean) { props.selected = s },
		unmount: () => (unmount(component)),
	};
};
type TimelineMilestoneElement = ReturnType<typeof createTimelineMilestoneElement>;

const getBackgroundColorForMilestoneKind = (kind: IncidentMilestone["attributes"]["kind"]) => {
	switch (kind) {
		case "impact": return "background-color: #f6ad55;";
		case "detection": return "background-color:rgb(74, 163, 144);";
		case "investigation": return "background-color:rgb(107, 39, 149);";
		case "mitigation": return "background-color:rgb(136, 186, 61);";
		case "resolution": return "background-color: #48bb78;";
	}
}

const getBackgroundStylesForMilestone = (m: IncidentMilestone) => {
	return "opacity: 0.15;" + getBackgroundColorForMilestoneKind(m.attributes.kind);
}

const getMilestoneIds = (id: string) => ({bg: id + "_bg", box: id + "_box"});

const makeMilestoneTimelineItems = (el: TimelineMilestoneElement, ms: IncidentMilestone, endTime: Date) => {
	const start = new Date(ms.attributes.timestamp);

	const ids = getMilestoneIds(ms.id);

	const bgItem: TimelineItem = {
		id: ids.bg,
		type: "background",
		content: "",
		style: getBackgroundStylesForMilestone(ms),
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

const sortMilestones = (a: IncidentMilestone, b: IncidentMilestone) => {
	const aTs = new Date(a.attributes.timestamp).valueOf();
	const bTs = new Date(b.attributes.timestamp).valueOf();
	return aTs - bTs;
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
	milestones = $derived(this.milestonesQuery?.data?.data ?? []);

	onMilestonesDataUpdated() {
		const dataMap = new Map(this.milestones.map(m => [m.id, m]));
		const removeIds: string[] = [];

		this.items?.forEach((item, rawId) => {
			if (item.subgroup !== "milestones") return;
			const msId = String(rawId).split("_")[0];
			if (!dataMap.get(msId)) removeIds.push(msId);
		});

		removeIds.forEach(this.removeMilestone);
		this.milestones.toSorted(sortMilestones).forEach((ms, idx, arr) => {
			this.setMilestone(ms, idx < arr.length - 1 ? arr[idx + 1] : undefined);
		});

		if (this.items?.flush) this.items.flush();
	};

	setMilestone(ms: IncidentMilestone, nextMs?: IncidentMilestone) {
		let el = this.timelineElements.get(ms.id);
		if (!el) {
			el = createTimelineMilestoneElement(ms);
			this.timelineElements.set(ms.id, el);
		}

		const endDate = nextMs ? new Date(nextMs.attributes.timestamp) : this.incidentEnd;
		const msItems = makeMilestoneTimelineItems(el, ms, endDate);
		if (!!this.items?.get(msItems[0].id)) {
			this.items.update(msItems);
		} else {
			this.items?.add(msItems);
		}
	}

	setSelected(id: string, selected: boolean) {
		const msId = id.split("_")[0];
		const el = this.timelineElements.get(msId);
		if (!el) return;
		el.selected = selected;
	}

	removeMilestone(id: string) {
		const el = this.timelineElements.get(id);
		if (el) el.unmount();
		const ids = getMilestoneIds(id);
		this.items?.remove([ids.bg, ids.box]);
		this.timelineElements.delete(id);
	}
}

export class TimelineState {
	viewState = useIncidentViewState();

	items = new DataSet<TimelineItem>([]);
	events = new TimelineEventsState(this.items);
	milestones = new TimelineMilestonesState(this.items);

	timeline = $state<Timeline>();
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

			const openedAt = new Date(inc.attributes.openedAt);
			const closedAt = new Date(inc.attributes.closedAt);
	
			this.items.clear();
			this.items.add({ id: "incidentStart", type: "box", group: "default", subgroup: "incident", align: "left", content: "Incident Opened", start: openedAt });
			this.items.add({ id: "incidentClosed", type: "box", group: "default", subgroup: "incident", align: "right", content: "Incident Closed", start: closedAt });
			if (this.items.flush) this.items.flush();
			
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

		this.timeline.on("select", e => this.onTimelineSelect(e));

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