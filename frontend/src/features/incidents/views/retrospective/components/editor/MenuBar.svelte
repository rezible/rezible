<script lang="ts">
	import { Button, Tooltip, Icon, Menu, MenuItem, Toggle } from 'svelte-ux';
	import {
		mdiBug,
		mdiCodeBlockTags,
		mdiCodeTags,
		mdiFormatBold,
		mdiFormatItalic,
		mdiFormatListBulleted,
		mdiFormatListNumbered,
		mdiFormatQuoteOpen,
		mdiText, mdiFormatHeader1, mdiFormatHeader2, mdiChevronDown,
        mdiFormatListCheckbox
	} from '@mdi/js';
	import { activeEditor, activeStatus } from '../../lib/editor.svelte';

	const getIconForStatus = () => {
		if (activeStatus.paragraph) return mdiText;
		if (activeStatus.heading1) return mdiFormatHeader1;
		if (activeStatus.heading2) return mdiFormatHeader2;
		return mdiText;
	}
	// TODO: make sure this ^ doesn't reset when editor unfocused (frozen state?)

	const runCmd = $derived(activeEditor.tryRunCommand);
</script>

<div class="flex items-center w-full divide-x divide-surface-100">
	{#snippet formatMenuItem(name: string, active: boolean, icon: string, cmd: VoidFunction)}
		<MenuItem
			{icon}
			on:click={() => cmd()}
			selected={active}
		>
			{name}
		</MenuItem>
	{/snippet}

	<Toggle let:on={open} let:toggle let:toggleOff>
		<Button
			icon={getIconForStatus()}
			on:click={toggle}
			classes={{ root: 'px-2' }}
			variant={open ? 'fill' : 'fill-light'}
			rounded={false}
		>
			<Icon data={mdiChevronDown} />

			<Menu {open} on:close={toggleOff}>
				{@render formatMenuItem("Regular Text", activeStatus.paragraph, mdiText, runCmd(c => c.setParagraph()))}
				{@render formatMenuItem("Heading", activeStatus.heading1, mdiFormatHeader1, runCmd(c => c.toggleHeading({level: 1})))}
				{@render formatMenuItem("Subheading", activeStatus.heading2, mdiFormatHeader2, runCmd(c => c.toggleHeading({level: 2})))}
			</Menu>
		</Button>
	</Toggle>

	{#snippet markButton(tooltip: string, active: boolean, icon: string, cmd: VoidFunction)}
		<Tooltip title={tooltip}>
			<Button
				icon={icon}
				rounded={false}
				variant={active ? 'fill' : 'fill-light'}
				on:click={() => cmd()}
			/>
		</Tooltip>
	{/snippet}

	<div class="px-2">
		{@render markButton("Bold", activeStatus.bold, mdiFormatBold, runCmd(cmd => cmd.toggleBold()))}
		{@render markButton("Italic", activeStatus.italic, mdiFormatItalic, runCmd(cmd => cmd.toggleItalic()))}
		{@render markButton("Code", activeStatus.code, mdiCodeTags, runCmd(cmd => cmd.toggleCode()))}
	</div>

	<div class="px-2">
		{@render markButton("Code Block", activeStatus.codeBlock, mdiCodeBlockTags, runCmd(cmd => cmd.toggleCodeBlock()))}
		{@render markButton("Quote", activeStatus.blockquote, mdiFormatQuoteOpen, runCmd(cmd => cmd.toggleBlockquote()))}
	</div>

	<div class="px-2">
		{@render markButton("Numbered List", activeStatus.orderedList, mdiFormatListNumbered, runCmd(cmd => cmd.toggleOrderedList()))}
		{@render markButton("Bullet List", activeStatus.bulletList, mdiFormatListBulleted, runCmd(cmd => cmd.toggleBulletList()))}
		{@render markButton("Task List", activeStatus.taskList, mdiFormatListCheckbox, runCmd(cmd => cmd.toggleTaskList()))}
	</div>

	<!--Button
		icon={mdiBug}
		rounded={false}
		on:click={() => {
			if (activeEditor.editor) console.log(activeEditor.editor.getJSON());
		}}
	/-->
</div>
