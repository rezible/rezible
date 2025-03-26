<!-- TODO: just use github.com/bherbruck/svelte-echarts when svelte 5 released -->

<!--
The MIT License (MIT)

Copyright (c) 2016-present GU Yiling & ECOMFE

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
-->

<script lang="ts" module>
	type OmitHandlers<T> = {
		[K in keyof T as K extends `on${string}` ? never : K]: T[K];
	};

	export type ChartProps = {
		init: typeof baseInit | typeof coreInit;
		options: EChartsOption;
		theme?: "light" | "dark" | object;
		initOptions?: EChartsInitOpts;
		notMerge?: SetOptionOpts["notMerge"];
		lazyUpdate?: SetOptionOpts["lazyUpdate"];
		silent?: SetOptionOpts["silent"];
		replaceMerge?: SetOptionOpts["replaceMerge"];
		transition?: SetOptionOpts["transition"];
		chart?: BaseEchartsType | CoreEchartsType;
	} & EventHandlers &
		OmitHandlers<HTMLAttributes<HTMLDivElement>>;

	// ref: https://echarts.apache.org/en/api.html#events
	export const MOUSE_EVENT_NAMES = [
		'click',
		'dblclick',
		'mousedown',
		'mousemove',
		'mouseover',
		'mouseout',
		'globalout',
		'contextmenu',
	] as const

	export const INTERACTION_EVENT_NAMES = [
		'highlight',
		'downplay',
		'selectchanged',
		'legendselectchanged',
		'legendselected',
		'legendunselected',
		'legendinverseselect',
		'legendscroll',
		'datazoom',
		'datarangeselected',
		'timelinechanged',
		'timelineplaychanged',
		'restore',
		'dataviewchanged',
		'magictypechanged',
		'geoselectchanged',
		'geoselected',
		'geounselected',
		'axisareaselected',
		'brush',
		'brushend',
		'brushselected',
		'globalcursortaken',
		'rendered',
		'finished',
	] as const

	export const EVENT_NAMES = [...MOUSE_EVENT_NAMES, ...INTERACTION_EVENT_NAMES]

	export type ECMouseEvent = CallbackDataParams & {
		onevent?(event: MouseEvent): void
	}

	export type ECInteractionEvent = CallbackDataParams

	// event dispatch types don't work unless I manually do this???
	export type MouseEventHandlers = {
		onclick?(event: ECMouseEvent): void
		ondblclick?(event: ECMouseEvent): void
		onmousedown?(event: ECMouseEvent): void
		onmousemove?(event: ECMouseEvent): void
		onmouseover?(event: ECMouseEvent): void
		onmouseout?(event: ECMouseEvent): void
		onglobalout?(event: ECMouseEvent): void
		oncontextmenu?(event: ECMouseEvent): void
	}

	export type InteractionEventHandlers = {
		onhighlight?(event: ECInteractionEvent): void
		ondownplay?(event: ECInteractionEvent): void
		onselectchanged?(event: ECInteractionEvent): void
		onlegendselectchanged?(event: ECInteractionEvent): void
		onlegendselected?(event: ECInteractionEvent): void
		onlegendunselected?(event: ECInteractionEvent): void
		onlegendinverseselect?(event: ECInteractionEvent): void
		onlegendscroll?(event: ECInteractionEvent): void
		ondatazoom?(event: ECInteractionEvent): void
		ondatarangeselected?(event: ECInteractionEvent): void
		ontimelinechanged?(event: ECInteractionEvent): void
		ontimelineplaychanged?(event: ECInteractionEvent): void
		onrestore?(event: ECInteractionEvent): void
		ondataviewchanged?(event: ECInteractionEvent): void
		onmagictypechanged?(event: ECInteractionEvent): void
		ongeoselectchanged?(event: ECInteractionEvent): void
		ongeoselected?(event: ECInteractionEvent): void
		ongeounselected?(event: ECInteractionEvent): void
		onaxisareaselected?(event: ECInteractionEvent): void
		onbrush?(event: ECInteractionEvent): void
		onbrushend?(event: ECInteractionEvent): void
		onbrushselected?(event: ECInteractionEvent): void
		onglobalcursortaken?(event: ECInteractionEvent): void
		onrendered?(event: ECInteractionEvent): void
		onfinished?(event: ECInteractionEvent): void
	}

	export type EventHandlers = MouseEventHandlers & InteractionEventHandlers

</script>

<script lang="ts">
	import type {
		init as baseInit,
		EChartsType as BaseEchartsType,
		EChartsOption,
		SetOptionOpts,
	} from "echarts";
	import type { CallbackDataParams } from 'echarts/types/dist/shared.js'
	import type { init as coreInit, EChartsType as CoreEchartsType } from "echarts/core";
	import type { EChartsInitOpts } from "echarts";
	import { onMount } from "svelte";
	import type { HTMLAttributes } from "svelte/elements";

	let {
		init,
		theme = "light",
		initOptions = {},
		options,
		notMerge = true,
		lazyUpdate = false,
		silent = false,
		replaceMerge,
		transition,
		chart = $bindable(),
		...restProps
	}: ChartProps = $props();

	// restProps is currently broken with typescript
	const tsProps = $derived(
		Object.keys(restProps)
			.filter((key) => !key.startsWith("on"))
			.reduce((r, k) => ({ ...r, [k]: (restProps as any)[k] }), {} as HTMLAttributes<HTMLDivElement>)
	);

	let element: HTMLDivElement;

	const updateOptions = () => {
		chart?.setOption(options, { notMerge, lazyUpdate, silent, replaceMerge, transition });
	};
	$effect(updateOptions);

	const mountChart = () => {
		chart?.dispose();
		chart = init(element, theme, initOptions);

		EVENT_NAMES.forEach((eventName) => {
			// @ts-ignore
			chart!.on(eventName, (event) => {
				// @ts-ignore
				restProps["on" + eventName]?.(event);
			});
		});

		const resizeObserver = new ResizeObserver(() => chart?.resize());
		resizeObserver.observe(element);
		return () => {
			resizeObserver.disconnect();
			chart?.dispose();
		};
	};
	onMount(mountChart);
</script>

<div bind:this={element} style="width: 100%; height: 100%; {tsProps.style}" {...tsProps}></div>
