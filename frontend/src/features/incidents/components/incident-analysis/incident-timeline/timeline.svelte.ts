import { mount, onMount, unmount } from "svelte";
import { Timeline, type DataGroup, type TimelineItem, type TimelineOptions } from "vis-timeline/esnext";
import { DataSet } from "vis-data/esnext";

import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Previous, watch } from "runed";
import { incidentCtx } from "$features/incidents/lib/context.ts";
import { listIncidentEventsOptions, listIncidentMilestonesOptions, type IncidentEvent, type IncidentMilestone } from "$lib/api";

import IncidentTimelineEventItemContent, { type Props as TimelineEventComponentProps } from "./IncidentTimelineEventItemContent.svelte";
import IncidentTimelineMilestoneItemContent, { type Props as TimelineMilestoneComponentProps } from "./IncidentTimelineMilestoneItemContent.svelte";

import { debounce } from "$src/lib/utils.svelte";

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
type TimelineEventElement = ReturnType<typeof createTimelineEventElement>;

const EventsSubgroup = "events";
const MilestonesSubgroup = "milestones";

const fitTimeline = debounce((t: Timeline) => { t?.fit({ zoom: true, animation: { duration: 500, easingFunction: "easeInOutQuad" } }) }, 100);

const createTimelineEventsState = (items: DataSet<TimelineItem>) => {
	let timeline: Timeline;
	const incident = incidentCtx.get();
	const queryClient = useQueryClient();

	const listEventsQueryOpts = listIncidentEventsOptions({ path: { id: incident.id } });
	const eventsQuery = createQuery(() => listEventsQueryOpts);

	const events = $derived(eventsQuery?.data?.data ?? []);

	const timelineElements = new Map<string, TimelineEventElement>();

	const makeTimelineEventItem = (el: TimelineEventElement, event: IncidentEvent) => {
		const start = new Date(event.attributes.timestamp);
		return { id: event.id, type: "box", group: "default", subgroup: EventsSubgroup, content: el.ref, start } as TimelineItem;
	}

	const setEvent = (event: IncidentEvent) => {
		const curr = timelineElements.get(event.id);
		if (curr /* && needsUpdate */) {
			items.update(makeTimelineEventItem(curr, event));
		} else {
			const el = createTimelineEventElement(event);
			timelineElements.set(event.id, el);
			items.add(makeTimelineEventItem(el, event));
		}
	}

	const removeEvent = (id: string) => {
		const el = timelineElements.get(id);
		if (el) el.unmount();
		items.remove(id);
		timelineElements.delete(id);
	}

	const onEventsDataUpdated = () => {
		const eventsMap = new Map(events.map(ev => [ev.id, ev]));
		const removeIds: string[] = [];

		items.forEach((item, rawId) => {
			if (item.subgroup !== "events") return;
			const id = String(rawId);
			if (eventsMap.has(id)) return;
			removeIds.push(id);
		});

		removeIds.forEach(removeEvent);
		eventsMap.forEach(setEvent);

		if (items.flush) items.flush();

		fitTimeline(timeline);
	};
	watch(() => events, onEventsDataUpdated);

	const eventAdded = (event: IncidentEvent) => {
		queryClient.setQueryData(listEventsQueryOpts.queryKey, current => {
			if (!current) return current;
			return { ...current, data: [...current.data, event] };
		});
		queryClient.invalidateQueries(listEventsQueryOpts);
	};

	const setSelected = (id: string, selected: boolean) => {
		const el = timelineElements.get(id);
		if (!el) return;
		el.props.selected = selected;
	}

	const clear = () => {
		timelineElements.forEach(c => c.unmount());
		timelineElements.clear();
	}

	const setTimeline = (t: Timeline) => {
		timeline = t;
	}

	return {
		eventAdded,
		setSelected,
		setTimeline,
		get eventsData() { return events },
		clear,
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

const createTimelineMilestonesState = (items: DataSet<TimelineItem>) => {
	const incident = incidentCtx.get();
	const incidentEnd = new Date(incident.attributes.closedAt);

	let timeline: Timeline;

	const listMilestonesQueryOpts = $derived(listIncidentMilestonesOptions({ path: { id: incident.id } }));
	const milestonesQuery = createQuery(() => listMilestonesQueryOpts);
	const milestones = $derived(milestonesQuery?.data?.data ?? []);

	const timelineElements = new Map<string, TimelineMilestoneElement>();

	const getIds = (id: string) => ({bg: id + "_bg", box: id + "_box"});

	const makeTimelineItems = (el: TimelineMilestoneElement, ms: IncidentMilestone, nextMs?: IncidentMilestone) => {
		const start = new Date(ms.attributes.timestamp);
		const end = nextMs ? new Date(nextMs.attributes.timestamp) : incidentEnd;

		const ids = getIds(ms.id);

		const bgItem: TimelineItem = {
			id: ids.bg,
			type: "background",
			content: "",
			style: getBackgroundStylesForMilestone(ms),
			start,
			end,
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

	const setMilestone = (ms: IncidentMilestone, nextMs?: IncidentMilestone) => {
		let el = timelineElements.get(ms.id);
		if (!el) {
			el = createTimelineMilestoneElement(ms);
			timelineElements.set(ms.id, el);
		}

		const msItems = makeTimelineItems(el, ms, nextMs);
		if (!!items.get(msItems[0].id)) {
			items.update(msItems);
		} else {
			items.add(msItems);
		}
	}

	const removeMilestone = (id: string) => {
		const el = timelineElements.get(id);
		if (el) el.unmount();
		const ids = getIds(id);
		items.remove([ids.bg, ids.box]);
		timelineElements.delete(id);
	}

	const sortMilestones = (a: IncidentMilestone, b: IncidentMilestone) => {
		const aTs = new Date(a.attributes.timestamp).valueOf();
		const bTs = new Date(b.attributes.timestamp).valueOf();
		return aTs - bTs;
	}

	const onMilestonesDataUpdated = () => {
		const dataMap = new Map(milestones.map(m => [m.id, m]));
		const removeIds: string[] = [];

		items.forEach((item, rawId) => {
			if (item.subgroup !== "milestones") return;
			const msId = String(rawId).split("_")[0];
			if (!dataMap.get(msId)) removeIds.push(msId);
		});

		removeIds.forEach(removeMilestone);
		milestones.toSorted(sortMilestones).forEach((ms, idx, arr) => {
			setMilestone(ms, idx < arr.length - 1 ? arr[idx + 1] : undefined);
		});

		if (items.flush) items.flush();

		fitTimeline(timeline);
	};
	watch(() => milestones, onMilestonesDataUpdated);

	const clear = () => {
		timelineElements.forEach(c => c.unmount());
		timelineElements.clear();
	}

	const setSelected = (id: string, selected: boolean) => {
		const msId = id.split("_")[0];
		const el = timelineElements.get(msId);
		if (!el) return;
		el.selected = selected;
	}

	const setTimeline = (t: Timeline) => {
		timeline = t;
	}

	return {
		clear,
		setSelected,
		setTimeline,
		get milestonesData() { return milestones },
	}
}

const createTimelineState = () => {
	const items = $state(new DataSet<TimelineItem>([]));
	let events = $state<ReturnType<typeof createTimelineEventsState>>();
	let milestones = $state<ReturnType<typeof createTimelineMilestonesState>>();

	let timeline = $state<Timeline>();
	let selectedItems = $state<Set<string>>(new Set());

	const onTimelineSelect = (e: any) => {
		const newSelected = new Set(e.items as string[]);
		const deselectedItems = selectedItems.difference(newSelected);
		
		deselectedItems.forEach(id => {
			events?.setSelected(id, false);
			milestones?.setSelected(id, false);
		});
		newSelected.forEach(id => {
			events?.setSelected(id, true);
			milestones?.setSelected(id, true);
		});

		selectedItems = newSelected;
	}

	const setup = (containerRefFn: () => HTMLElement | undefined) => {
		const incident = incidentCtx.get();

		events = createTimelineEventsState(items);
		milestones = createTimelineMilestonesState(items);

		const openedAt = new Date(incident.attributes.openedAt);
		const closedAt = new Date(incident.attributes.closedAt);

		items.clear();

		items.add({ id: "incidentStart", type: "box", group: "default", subgroup: "incident", align: "left", content: "Incident Opened", start: openedAt });
		items.add({ id: "incidentClosed", type: "box", group: "default", subgroup: "incident", align: "right", content: "Incident Closed", start: closedAt });

		const timelineOpts: TimelineOptions = {
			height: "100%",

			zoomMin: 1000 * 60 * 60,
			zoomMax: 1000 * 60 * 60 * 24,
		};

		const timelineGroups: DataGroup[] = [
			{ id: "default", title: "Incident", content: "", subgroupStack: { "incident": true } },
		];

		watch(containerRefFn, ref => {
			if (!ref) return;
			// if (timeline) timeline.destroy();
			// @ts-expect-error incorrect timeline DataItem typing for content
			timeline = new Timeline(ref, items, timelineGroups, timelineOpts);
			events?.setTimeline(timeline);
			milestones?.setTimeline(timeline);

			timeline.on("select", onTimelineSelect)
		});
		onMount(() => {
			return () => {
				timeline?.destroy();
				events?.clear();
				milestones?.clear();
				items.clear();
			};
		});
	};

	const onEventAdded = (event: IncidentEvent) => events?.eventAdded(event);

	return {
		setup,
		onEventAdded,
		get milestones() { return milestones?.milestonesData ?? [] },
		get events() { return events?.eventsData ?? [] },
	};
};
export const timeline = createTimelineState();
