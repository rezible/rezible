<script lang="ts">
	import RiSlackLine from "remixicon-svelte/icons/slack-line";
	import TeamUsersTable from "./TeamUsersTable.svelte";
	import TeamRostersTable from "./TeamRostersTable.svelte";
	import { initTeamOverviewController } from "./controller.svelte";

	initTeamOverviewController();

	const slackChannelOptions = [{ label: "team-channel", value: "foo" }];
	const slackChannel = slackChannelOptions[0].value;
</script>

<div class="flex gap-2 p-2">
	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-foreground/90">Details</span>

		<span>slack channel</span>
		<!-- <SelectField
			label="Slack Channel"
			labelPlacement="top"
			icon={RiSlackLine}
			readonly
			options={slackChannelOptions}
			value={slackChannel}
			clearable={false}
		/> -->
	</div>

	<TeamUsersTable />

	<TeamRostersTable />
</div>
