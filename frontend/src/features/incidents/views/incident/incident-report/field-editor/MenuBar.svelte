<script lang="ts">
	import { Button, Tooltip, Menu, MenuItem, Toggle } from "svelte-ux";
	import {
		mdiCodeBlockTags,
		mdiCodeTags,
		mdiFormatBold,
		mdiFormatItalic,
		mdiFormatListBulleted,
		mdiFormatListNumbered,
		mdiFormatQuoteOpen,
		mdiText,
		mdiFormatHeader1,
		mdiFormatHeader2,
		mdiChevronDown,
		mdiFormatListCheckbox,
	} from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";
	import { activeEditor, activeStatus } from "$features/incidents/lib/activeEditor.svelte";

	const getIconForStatus = () => {
		if (activeStatus.paragraph) return mdiText;
		if (activeStatus.heading1) return mdiFormatHeader1;
		if (activeStatus.heading2) return mdiFormatHeader2;
		return mdiText;
	};
	// let lastFocusedStatus = $state()
	let formatIcon = $state(mdiText);
	$effect(() => {
		if (!activeStatus.focused) return;
		formatIcon = getIconForStatus();
	});
	// TODO: make sure this ^ doesn't reset when editor unfocused (frozen state?)

	const runCmd = $derived(activeEditor.tryRunCommand);
</script>

<div class="flex items-center w-full divide-x divide-surface-100 h-8">
	{#snippet formatMenuItem(name: string, active: boolean, icon: string, cmd: VoidFunction)}
		<MenuItem {icon} on:click={() => cmd()} selected={active}>
			{name}
		</MenuItem>
	{/snippet}

	<Toggle let:on={open} let:toggle let:toggleOff>
		<Button
			icon={formatIcon}
			on:click={toggle}
			classes={{ root: "px-2 h-8" }}
			variant={open ? "fill-light" : "text"}
			rounded={false}
		>
			<Icon data={mdiChevronDown} />

			<Menu {open} on:close={toggleOff}>
				{@render formatMenuItem(
					"Regular Text",
					activeStatus.paragraph,
					mdiText,
					runCmd((c) => c.setParagraph())
				)}
				{@render formatMenuItem(
					"Heading",
					activeStatus.heading1,
					mdiFormatHeader1,
					runCmd((c) => c.toggleHeading({ level: 1 }))
				)}
				{@render formatMenuItem(
					"Subheading",
					activeStatus.heading2,
					mdiFormatHeader2,
					runCmd((c) => c.toggleHeading({ level: 2 }))
				)}
			</Menu>
		</Button>
	</Toggle>

	{#snippet markButton(tooltip: string, active: boolean, icon: string, cmd: VoidFunction)}
		<Tooltip title={tooltip}>
			<Button
				classes={{ root: "size-8" }}
				{icon}
				rounded={false}
				color={active ? "secondary" : "default"}
				variant={active ? "fill-light" : "text"}
				on:click={() => cmd()}
			/>
		</Tooltip>
	{/snippet}

	<div class="px-2">
		{@render markButton(
			"Bold",
			activeStatus.bold,
			mdiFormatBold,
			runCmd((cmd) => cmd.toggleBold())
		)}
		{@render markButton(
			"Italic",
			activeStatus.italic,
			mdiFormatItalic,
			runCmd((cmd) => cmd.toggleItalic())
		)}
		{@render markButton(
			"Code",
			activeStatus.code,
			mdiCodeTags,
			runCmd((cmd) => cmd.toggleCode())
		)}
	</div>

	<div class="px-2">
		{@render markButton(
			"Code Block",
			activeStatus.codeBlock,
			mdiCodeBlockTags,
			runCmd((cmd) => cmd.toggleCodeBlock())
		)}
		{@render markButton(
			"Quote",
			activeStatus.blockquote,
			mdiFormatQuoteOpen,
			runCmd((cmd) => cmd.toggleBlockquote())
		)}
	</div>

	<div class="px-2">
		{@render markButton(
			"Numbered List",
			activeStatus.orderedList,
			mdiFormatListNumbered,
			runCmd((cmd) => cmd.toggleOrderedList())
		)}
		{@render markButton(
			"Bullet List",
			activeStatus.bulletList,
			mdiFormatListBulleted,
			runCmd((cmd) => cmd.toggleBulletList())
		)}
		{@render markButton(
			"Task List",
			activeStatus.taskList,
			mdiFormatListCheckbox,
			runCmd((cmd) => cmd.toggleTaskList())
		)}
	</div>

	<!--Button
		icon={mdiBug}
		rounded={false}
		on:click={() => {
			if (activeEditor.editor) console.log(activeEditor.editor.getJSON());
		}}
	/-->
</div>
