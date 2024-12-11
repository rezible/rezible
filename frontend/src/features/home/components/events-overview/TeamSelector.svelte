<script lang="ts">
    import Avatar from "$src/components/avatar/Avatar.svelte";
import { mdiAccountGroup, mdiGroup, mdiMagnify } from "@mdi/js";
    import { Icon, SelectField, type MenuOption } from "svelte-ux";

	type Props = {
		onChange: (id: string) => void;
	}
	const { onChange }: Props = $props();

	const teams: MenuOption<string>[] = [
		{ label: 'One', value: "one", group: "Pinned" },
		{ label: 'Two', value: "two", group: "Pinned" },
		{ label: 'Four', value: "three", group: "All" },
		{ label: 'Five', value: "four", group: "All" },
	];

	let selected = $state<MenuOption<string> | undefined>();
	let teamOptions: MenuOption[] = $state(teams);

	const searchTeams = async () => {
		
    }
</script>

<SelectField matchWidth bind:selected options={teamOptions} search={searchTeams} on:change={e => {onChange(e.detail.value)}} placeholder="Select a Team">
	<div slot="prepend" class="mr-2 size-6 flex items-center">
		{#if selected}
			<Avatar kind="team" id={selected.value} size={24} />
		{:else}
			<Icon data={mdiAccountGroup} size={24} />
		{/if}
	</div>
</SelectField>
