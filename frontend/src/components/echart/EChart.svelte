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
</script>

<script lang="ts">
	import type {
		init as baseInit,
		EChartsType as BaseEchartsType,
		EChartsOption,
		SetOptionOpts,
	} from "echarts";
	import type { init as coreInit, EChartsType as CoreEchartsType } from "echarts/core";
	import type { EChartsInitOpts } from "echarts";
	import { EVENT_NAMES, type EventHandlers } from "./events";
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
