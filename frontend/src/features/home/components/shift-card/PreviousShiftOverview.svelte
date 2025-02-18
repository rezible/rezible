<script lang="ts">
	import { Button, Card, Collapse, Header, Icon, ProgressCircle } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { fade } from "svelte/transition";
	import { mdiCircleMedium, mdiAlarmLight, mdiSleepOff, mdiFire, mdiClose, mdiChevronUp, mdiChevronDown } from "@mdi/js";
	import HandoverOverview from "./HandoverOverview.svelte";

	type Props = {
		expanded: boolean;
	};
	let { expanded = $bindable() }: Props = $props();
</script>

<div class="rounded-lg bg-success-900/10 p-2 pr-3 flex flex-col gap-2 min-h-0 w-full overflow-auto">
	<Collapse classes={{root: "overflow-y-auto flex flex-col overflow-x-hidden", content: "flex-1 flex flex-col min-h-0"}}>
		<div slot="trigger" class="flex-1 px-3 py-3">
			<div class="flex items-center gap-1">
				<Button href="/oncall/rosters/search" classes={{ root: "p-1" }}>
					<Avatar kind="user" id="user-id" size={24} />
					<span class="font-bold text-base ml-2">John Doe</span>
				</Button>
				<span>was the previous oncaller</span>
			</div>
		</div>
		<div class="px-3 pb-3 flex flex-col gap-2 overflow-y-auto" slot="default">
			<div
				class="grid grid-cols-3 divide-x-2 border gap-3 rounded-xl bg-surface-200/50 items-stretch justify-items-center"
			>
				<div class="flex w-full justify-start items-center gap-4 p-2">
					<Icon data={mdiAlarmLight} />
					<div class="flex flex-col">
						<span class="text-lg">18 Alerts</span>
						<span class="text-surface-content/75">Normal</span>
					</div>
				</div>

				<div class="flex w-full justify-center items-center gap-4 p-2">
					<div class="">
						<Icon data={mdiSleepOff} />
					</div>
					<div class="flex flex-col">
						<span class="text-lg">6 Night Alerts</span>
						<span class="text-warning-800">Above Average</span>
					</div>
				</div>

				<div class="flex w-full justify-center items-center gap-4 p-2">
					<div class="">
						<Icon data={mdiFire} />
					</div>
					<div class="flex flex-col">
						<span class="text-lg">2 Incidents</span>
						<span class="text-warning-800">Above Average</span>
					</div>
				</div>
			</div>

			<HandoverOverview />
		</div>
	</Collapse>
</div>
