<script lang="ts">
	import { Button } from "$components/ui/button";
	import type { Component } from "svelte";

	import RiCodeBoxLine from "remixicon-svelte/icons/code-box-line";
	import RiCodeLine from "remixicon-svelte/icons/code-line";
	import RiBold from "remixicon-svelte/icons/bold";
	import RiItalic from "remixicon-svelte/icons/italic";
	import RiListUnordered from "remixicon-svelte/icons/list-unordered";
	import RiListOrdered from "remixicon-svelte/icons/list-ordered";
	import RiDoubleQuotesL from "remixicon-svelte/icons/double-quotes-l";
	import RiText from "remixicon-svelte/icons/text";
	import RiH1 from "remixicon-svelte/icons/h-1";
	import RiH2 from "remixicon-svelte/icons/h-2";
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";
	import RiListCheck2 from "remixicon-svelte/icons/list-check-2";
	
	import { activeEditor, activeStatus } from "../activeEditor.svelte";

	const getIconForStatus = () => {
		if (activeStatus.paragraph) return RiText;
		if (activeStatus.heading1) return RiH1;
		if (activeStatus.heading2) return RiH2;
		return RiText;
	};
	// let lastFocusedStatus = $state()
	let formatIcon = $state(RiText);
	$effect(() => {
		if (!activeStatus.focused) return;
		formatIcon = getIconForStatus();
	});
	// TODO: make sure this ^ doesn't reset when editor unfocused (frozen state?)

	const runCmd = $derived(activeEditor.tryRunCommand);
</script>

<div class="flex items-center w-full divide-x divide-surface-100 h-8">
	{#snippet formatMenuItem(name: string, active: boolean, icon: string, cmd: VoidFunction)}
		<!-- <MenuItem {icon} onclick={() => cmd()} selected={active}>
			{name}
		</MenuItem> -->
	{/snippet}

	<!--Toggle let:on={open} let:toggle let:toggleOff>
		<Button
			icon={formatIcon}
			onclick={toggle}
			classes={{ root: "px-2 h-8" }}
			variant={open ? "fill-light" : "text"}
			rounded={false}
		>
			<RiArrowDownSLine class="" aria-hidden="true" />

			<Menu {open} on:close={toggleOff}>
				{@render formatMenuItem(
					"Regular Text",
					activeStatus.paragraph,
					RiText,
					runCmd((c) => c.setParagraph())
				)}
				{@render formatMenuItem(
					"Heading",
					activeStatus.heading1,
					RiH1,
					runCmd((c) => c.toggleHeading({ level: 1 }))
				)}
				{@render formatMenuItem(
					"Subheading",
					activeStatus.heading2,
					RiH2,
					runCmd((c) => c.toggleHeading({ level: 2 }))
				)}
			</Menu>
		</Button>
	</Toggle-->

	{#snippet markButton(tooltip: string, active: boolean, icon: Component, cmd: VoidFunction)}
		<Button color={active ? "secondary" : "default"} onclick={() => cmd()}>{tooltip}</Button>
	{/snippet}

	<div class="px-2">
		{@render markButton(
			"Bold",
			activeStatus.bold,
			RiBold,
			runCmd((cmd) => cmd.toggleMark("bold"))
		)}
		{@render markButton(
			"Italic",
			activeStatus.italic,
			RiItalic,
			runCmd((cmd) => cmd.toggleMark("italic"))
		)}
		{@render markButton(
			"Code",
			activeStatus.code,
			RiCodeLine,
			runCmd((cmd) => cmd.toggleMark("code"))
		)}
	</div>

	<!--
	<div class="px-2">
		{@render markButton(
			"Code Block",
			activeStatus.codeBlock,
			RiCodeBoxLine,
			runCmd((cmd) => cmd.toggleCodeBlock())
		)}
		{@render markButton(
			"Quote",
			activeStatus.blockquote,
			RiDoubleQuotesL,
			runCmd((cmd) => cmd.toggleBlockquote())
		)}
	</div>

	<div class="px-2">
		{@render markButton(
			"Numbered List",
			activeStatus.orderedList,
			RiListOrdered,
			runCmd((cmd) => cmd.toggleOrderedList())
		)}
		{@render markButton(
			"Bullet List",
			activeStatus.bulletList,
			RiListUnordered,
			runCmd((cmd) => cmd.toggleBulletList())
		)}
		{@render markButton(
			"Task List",
			activeStatus.taskList,
			RiListCheck2,
			runCmd((cmd) => cmd.toggleTaskList())
		)}
	</div>
	-->

	<!--Button
		icon={RiBugLine}
		rounded={false}
		onclick={() => {
			if (activeEditor.editor) console.log(activeEditor.editor.getJSON());
		}}
	/-->
</div>
