<script lang="ts">
	import { Button, Icon } from "svelte-ux";
	import { mdiArrowRight, mdiChevronLeft, mdiChevronRight } from "@mdi/js";
	import { session } from "$lib/auth.svelte";
	import { buildShiftTimeDetails } from "$features/oncall/lib/utils";
	import { page } from "$app/state";
	import { getOncallShiftOptions } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";

	const shiftId = $derived(page.params.id);
	const query = createQuery(() => getOncallShiftOptions({ path: { id: shiftId } }));
	const shift = $derived(query.data?.data);

	const previousShiftId = $derived(shiftId);
	const nextShiftId = $derived(shiftId);

	const isActiveShift = $derived(shift ? buildShiftTimeDetails(shift).status === "active" : false);
	const isSessionUser = $derived(session.userId == shift?.attributes.user.id);
</script>

<div class="flex gap-2">
	<Button href="/oncall/shifts/{previousShiftId}">
		<Icon data={mdiChevronLeft} />
		<span class="">Previous Shift</span>
	</Button>

	{#if shift && isActiveShift && isSessionUser}
		<Button
			href="/oncall/shifts/{shift.id}/handover"
			variant="fill"
			color="secondary"
		>
			<span>Edit Handover</span>
		</Button>
	{/if}

	<Button href="/oncall/shifts/{nextShiftId}">
		<span class="">Next Shift</span>
		<Icon data={mdiChevronRight} />
	</Button>
</div>