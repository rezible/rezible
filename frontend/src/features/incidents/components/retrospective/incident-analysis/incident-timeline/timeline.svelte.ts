import { mount, onMount, unmount } from "svelte";
import {
  Timeline,
  type IdType,
  type TimelineOptions,
} from "vis-timeline/esnext";
import { DataSet } from "vis-data/esnext";

import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { watch } from "runed";
import { incidentCtx } from "$features/incidents/lib/context.ts";
import {
  listIncidentMilestonesOptions,
  type IncidentMilestone,
  type ListIncidentMilestonesResponse,
} from "$lib/api";
import IncidentTimelineEvent, {
  type TimelineEventComponentProps,
} from "./IncidentTimelineEvent.svelte";
import type { TimelineEvent } from "./types";

export const createTimelineEventElement = (id: string) => {
  let props = $state<TimelineEventComponentProps>({ label: "initial" });

  const target = document.createElement("div");
  target.setAttribute("event-id", id);

  const component = mount(IncidentTimelineEvent, { target, props });

  return {
    get element() {
      return target;
    },
    setLabel: (label: string) => {
      props.label = label;
    },
    unmount: () => {
      unmount(component);
    },
  };
};

const createTimelineState = () => {
  let timeline = $state<Timeline>();

  let milestoneItems = new DataSet<any>([]);
  let eventItems = new DataSet<any>([]);

  const items = new DataSet<any>([]);
  const updateItems = () => {
    // items = new DataSet()
  };

  const onMilestonesQueryDataUpdated = (m: IncidentMilestone[]) => {
    milestoneItems = new DataSet([
      {
        id: "A",
        content: "Period A",
        start: "2014-01-16",
        end: "2014-01-22",
        type: "background",
      },
      {
        id: "B",
        content: "Period B",
        start: "2014-01-25",
        end: "2014-01-30",
        type: "background",
        className: "negative",
      },
    ]);
  };

  const onEventsQueryDataUpdated = (events: any[]) => {
    console.log("events updated");
  };

  const createQueries = () => {
    const queryClient = useQueryClient();
    const incidentId = incidentCtx.get().id;

    const milestonesQuery = createQuery(
      () => listIncidentMilestonesOptions({ path: { id: incidentId } }),
      queryClient,
    );
    watch(
      () => milestonesQuery,
      (r) => onMilestonesQueryDataUpdated(r.data?.data ?? []),
    );

    // TODO: swap this for correct query
    const eventsQuery = createQuery(
      () => listIncidentMilestonesOptions({ path: { id: incidentId } }),
      queryClient,
    );
    watch(
      () => eventsQuery,
      (r) => onEventsQueryDataUpdated(r.data?.data ?? []),
    );
  };

  const eventComponents = new Map<
    IdType,
    ReturnType<typeof createTimelineEventElement>
  >();
  const addEvent = (id: IdType) => {
    const created = createTimelineEventElement(id.toString());
    items.add({ id: 1, content: created.element, start: "2014-01-23" });
    eventComponents.set(id, created);
  };

  const mount = (container: HTMLElement) => {
    const options: TimelineOptions = {
      height: "100%",
    };
    timeline = new Timeline(container, items, options);

    addEvent("bleh");
  };

  const unmount = () => {
    timeline?.destroy();
    eventComponents.forEach((c) => c.unmount());
    eventComponents.clear();
    items.clear();
  };

  const componentSetup = (containerElFn: () => HTMLElement | undefined) => {
    watch(containerElFn, (el) => {
      if (el) mount(el);
    });
    onMount(() => {
      return unmount;
    });
    createQueries();
  };

  let editingEvent = $state<TimelineEvent>();

  return {
    componentSetup,
    get editingEvent() {
      return editingEvent;
    },
  };
};
export const timeline = createTimelineState();
