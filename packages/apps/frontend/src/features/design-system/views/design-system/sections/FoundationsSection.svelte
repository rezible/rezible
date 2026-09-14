<script lang="ts">
	import RiCheckLine from "remixicon-svelte/icons/check-line";
	import RiTimeLine from "remixicon-svelte/icons/time-line";
	import * as Card from "$components/ui/card";
	import { Badge } from "$components/ui/badge";

	const surfaces = [
		["Canvas", "bg-background"],
		["Reading", "bg-card"],
		["Overlay", "bg-popover"],
		["Feature", "bg-feature"],
		["Hover", "bg-accent"],
		["Selection", "bg-selection"],
		["Primary", "bg-primary"],
		["Pressed", "bg-control-pressed"],
	] as const;
</script>

<section class="flex flex-col gap-6">
	<Card.Root>
		<Card.Header>
			<Card.Title>Colour and surfaces</Card.Title>
		</Card.Header>
		<Card.Content class="flex flex-col gap-3">
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-8">
				{#each surfaces as surface (surface[0])}
					<div class="overflow-hidden rounded-lg border bg-card">
						<div class={`h-16 ${surface[1]}`}></div>
						<p class="border-t px-3 py-2 text-xs font-medium">{surface[0]}</p>
					</div>
				{/each}
			</div>
			<div class="flex flex-wrap gap-2">
				<Badge variant="success"><RiCheckLine data-icon="inline-start" />Healthy</Badge>
				<Badge variant="warning"><RiTimeLine data-icon="inline-start" />Watching</Badge>
				<Badge variant="danger">Degraded</Badge>
				<Badge variant="neutral">Unknown</Badge>
				<Badge variant="info">Information</Badge>
			</div>
		</Card.Content>
	</Card.Root>

	<div class="grid gap-4 lg:grid-cols-2">
		<Card.Root>
			<Card.Header><Card.Title>Typography and report measure</Card.Title></Card.Header>
			<Card.Content class="flex flex-col gap-3">
				<p class="text-[28px] font-semibold leading-9">Page headline</p>
				<p class="text-lg font-semibold leading-[26px]">Section heading</p>
				<p class="text-[15px] font-medium leading-[22px]">Evidence title</p>
				<p class="text-sm leading-[22px]">Body and table text carries operational detail.</p>
				<p class="text-xs leading-[18px] text-muted-foreground">Secondary metadata</p>
				<code class="w-fit rounded-md bg-muted px-2 py-1 font-mono text-xs">EVT-7F3A91</code>
				<p class="tabular-nums text-sm">14:16:03 UTC · 2m 34s · 1,248 events</p>
				<p class="max-w-[70ch] text-[15px] leading-6">
					Report prose stays within a readable 65–75 character measure for sustained analysis and
					review.
				</p>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header><Card.Title>Spacing and geometry</Card.Title></Card.Header>
			<Card.Content class="flex flex-col gap-4">
				<div class="flex items-end gap-3">
					{#each [4, 8, 12, 16, 24, 32] as size}
						<div class="text-center text-xs text-muted-foreground">
							<div
								class="mx-auto mb-1 rounded-sm bg-brand"
								style={`width:${size}px;height:${size}px`}
							></div>
							{size}
						</div>
					{/each}
				</div>
				<div class="grid grid-cols-3 gap-3">
					<div class="h-9 rounded-md border p-2 text-xs">36px control</div>
					<div class="h-11 rounded-md border p-2 text-xs">44px target</div>
					<div class="min-h-[60px] rounded-lg border p-2 text-xs">60px evidence</div>
				</div>
				<div
					class="rounded-lg border bg-popover p-3 text-sm shadow-[0_4px_16px_rgb(28_34_37_/_0.10)] dark:shadow-[0_8px_24px_rgb(0_0_0_/_0.35)]"
				>
					Restrained overlay elevation
				</div>
			</Card.Content>
		</Card.Root>
	</div>
</section>
