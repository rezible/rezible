<script lang="ts">
	import type { RetrospectiveReportSection } from "$lib/api";
	import { mdiCircleMedium } from "@mdi/js";
	import { onMount } from "svelte";
	import Icon from "$components/icon/Icon.svelte";
	import { cls } from '@layerstack/tailwind';

	type Props = {
		visible: boolean;
		containerEl: HTMLElement;
		sections: RetrospectiveReportSection[];
		sectionElements: Record<string, HTMLElement>;
		onSectionClicked: (field: string) => void;
	};
	let { visible = $bindable(), containerEl, sections, sectionElements, onSectionClicked }: Props = $props();

	let progressBarContainerEl = $state<HTMLElement>();

	let scrollTop = $state(containerEl.scrollTop);
	let scrollHeight = $state(containerEl.scrollHeight);
	let clientHeight = $state(containerEl.clientHeight);
	let progessBarContainerHeight = $state(containerEl.clientHeight);

	// TODO: https://runed.dev/docs/utilities/element-size

	const updateHeights = () => {
		scrollHeight = containerEl.scrollHeight;
		clientHeight = containerEl.clientHeight;
		progessBarContainerHeight = progressBarContainerEl?.scrollHeight || clientHeight;
		const isScrolling = scrollHeight > clientHeight;
		if (isScrolling != visible) visible = isScrolling;
	};

	const windowHeight = $derived(progessBarContainerHeight * (clientHeight / scrollHeight));
	const scrollAmount = $derived(100 * ((clientHeight + scrollTop) / scrollHeight));
	const sectionSizes = $derived(
		sections.map((s) => (sectionElements[s.field]?.scrollHeight ?? 0) / scrollHeight)
	);

	const observer = new ResizeObserver(updateHeights);
	$effect(() => {
		observer.unobserve(containerEl);
		Object.values(sectionElements).forEach((el) => observer.observe(el));
		observer.observe(containerEl);
		return () => observer.disconnect();
	});

	onMount(() => {
		updateHeights();
		const updateScrollTop = () => {
			scrollTop = containerEl.scrollTop;
		};
		containerEl.addEventListener("scroll", updateScrollTop);
		return () => containerEl.removeEventListener("scroll", updateScrollTop);
	});
</script>

<div class="flex flex-col h-full items-start" class:hidden={!visible}>
	<!-- wrapper -->
	<div class="flex flex-row grow justify-evenly pl-1">
		<!-- progress bar container -->
		<div
			class="w-[1px] flex flex-col justify-evenly bg-surface-content/25 overflow-y-clip -mr-2"
			style="--scrollAmount: {scrollAmount}%; --windowHeight: {windowHeight}px"
			bind:this={progressBarContainerEl}
		>
			<!-- progress bar filled -->
			<div
				class={cls(
					"flex flex-[var(--scrollAmount)]",
					"after:h-[var(--windowHeight)] after:self-end after:bg-surface-content/25 after:content-[''] after:w-[1px] after:ml-[-0.5px] pl-[0.5px]"
				)}
			></div>

			<!-- progress bar unfilled -->
			<div style="width: 1px; flex: calc(100% - var(--scrollAmount))"></div>
		</div>
		<!-- elements -->
		<div class="flex flex-col grow">
			{#each sections as section, i}
				<!-- row wrapper -->
				<div
					class="flex relative flex-col-reverse z-1"
					style={i > 0 ? `flex: ${sectionSizes[i - 1]} 1 0%` : ""}
				>
					<!-- dot container -->
					<!--div class="grid grid-cols-2 items-center"-->
					<div class="flex flex-row">
						<!-- dot -->
						<div class="text-surface-content z-1 h-6 w-6 font-sm ml-[2px]">
							{#if i < 0}
								<Icon data={mdiCircleMedium} size={11} />
							{/if}
						</div>
						<div
							class="relative flex flex-col-reverse pl-0"
							class:text-primary-content={i === -1}
						>
							<button class="text-left" onclick={() => onSectionClicked(section.field)}
								>{section.title}</button
							>
						</div>
					</div>
				</div>
			{/each}
		</div>
	</div>
</div>
