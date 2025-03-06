<script lang="ts">
	import { Button, Icon } from "svelte-ux";
	import { mdiArrowRight } from "@mdi/js";
	import { session } from "$lib/auth.svelte";
	import { buildShiftTimeDetails } from "$features/oncall/lib/shift";
	import { page } from "$app/state";
	import { getOncallShiftOptions, type OncallShift } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";

	const shiftId = $derived(page.params.id);
	const query = createQuery(() => getOncallShiftOptions({ path: { id: shiftId } }));
	const shift = $derived(query.data?.data);

	const isActiveShift = $derived(shift ? buildShiftTimeDetails(shift).status === "active" : false);
	const isSessionUser = $derived(session.userId == shift?.attributes.user.id);
</script>

{#if shift && isActiveShift && isSessionUser}
	<Button
		href="/oncall/shifts/{shift.id}/handover"
		variant="fill"
		color="secondary"
	>
		<span>Edit Handover</span>
		<Icon data={mdiArrowRight} />
	</Button>
{/if}