<script lang="ts">
	import RiCheckLine from "remixicon-svelte/icons/check-line";
	import RiCloseLine from "remixicon-svelte/icons/close-line";
	import RiLoader4Line from "remixicon-svelte/icons/loader-4-line";
	import RiSearchLine from "remixicon-svelte/icons/search-line";
	import * as Dialog from "$components/ui/dialog/index.js";
	import { Badge } from "$components/ui/badge/index.js";
	import { Button } from "$components/ui/button/index.js";
	import { Input } from "$components/ui/input/index.js";
	import * as Table from "$components/ui/table/index.js";
	import * as Sidebar from "$components/ui/sidebar/index.js";
	import * as Tabs from "$components/ui/tabs/index.js";
	import { Toggle } from "$components/ui/toggle/index.js";
	import { Checkbox } from "$components/ui/checkbox/index.js";
	import * as RadioGroup from "$components/ui/radio-group/index.js";
	import { Switch } from "$components/ui/switch/index.js";

	let dialogOpen = $state(false);
	let checkboxChecked = $state(true);
	let switchChecked = $state(true);
</script>

<svelte:head><title>Visual identity specimen</title></svelte:head>

<div class="min-h-full overflow-auto bg-background p-6 text-foreground sm:p-8">
	<div class="mx-auto max-w-6xl space-y-8">
		<header class="border-b border-border pb-6">
			<p class="text-muted-foreground text-[11px] font-semibold uppercase tracking-[0.04em]">
				UI identity
			</p>
			<h1 class="mt-2 text-lg font-semibold leading-6">Visual identity specimen</h1>
			<p class="text-muted-foreground mt-2 max-w-2xl text-sm">
				Illustrative states for shared controls, semantic status, data density and overlay surfaces.
			</p>
		</header>

		<section class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
			<div class="space-y-6">
				<div class="border-border bg-card rounded-md border p-5">
					<h2 class="text-sm font-semibold">Actions and statuses</h2>
					<div class="mt-4 flex flex-wrap items-center gap-2">
						<Button><RiCheckLine />Open review</Button>
						<Button variant="outline">View evidence</Button>
						<Button variant="ghost">More</Button>
						<Button variant="destructive">Delete</Button>
						<Button disabled>Disabled</Button>
						<Button aria-label="Loading" class="min-w-28"
							><RiLoader4Line class="animate-spin" />Loading</Button
						>
					</div>
					<div class="mt-5 flex flex-wrap gap-2">
						<Badge>Neutral default</Badge>
						<Badge variant="success">Resolved</Badge>
						<Badge variant="warning">Watching</Badge>
						<Badge variant="danger">Degraded</Badge>
						<Badge variant="neutral">Unknown</Badge>
						<Badge variant="info">Information</Badge>
					</div>
				</div>

				<div class="border-border bg-card rounded-md border p-5">
					<h2 class="text-sm font-semibold">Inputs and focus</h2>
					<div class="mt-4 grid gap-4 sm:grid-cols-2">
						<label class="space-y-1.5 text-sm font-medium">
							Search incidents
							<div class="relative">
								<RiSearchLine
									class="text-muted-foreground pointer-events-none absolute top-2.5 left-2.5 size-4"
								/>
								<Input class="pl-9" value="Checkout search timeouts" />
							</div>
						</label>
						<label class="space-y-1.5 text-sm font-medium">
							Long label
							<Input placeholder="A readable placeholder remains visible" />
						</label>
						<label class="space-y-1.5 text-sm font-medium">
							Error state
							<Input aria-invalid="true" value="" placeholder="Incident title" />
							<span class="text-status-danger-foreground text-xs">A title is required</span>
						</label>
						<label class="space-y-1.5 text-sm font-medium">
							Disabled field
							<Input disabled value="Read-only value" />
						</label>
					</div>
				</div>

				<div class="border-border bg-card rounded-md border p-5">
					<h2 class="text-sm font-semibold">Selection and focus</h2>
					<div class="mt-4 flex flex-wrap items-center gap-3">
						<Toggle variant="outline" pressed>Toggle on</Toggle>
						<label class="flex items-center gap-2 text-sm"
							><Checkbox bind:checked={checkboxChecked} />Checkbox</label
						>
						<RadioGroup.Root value="selected" class="flex items-center gap-2">
							<RadioGroup.Item value="selected" aria-label="Selected radio" />
							<span class="text-sm">Radio</span>
						</RadioGroup.Root>
						<label class="flex items-center gap-2 text-sm"
							><Switch bind:checked={switchChecked} />Switch</label
						>
					</div>
					<div class="mt-5 space-y-3">
						<Tabs.Root value="filled">
							<Tabs.List>
								<Tabs.Trigger value="filled">Filled tab</Tabs.Trigger>
								<Tabs.Trigger value="line">Line tab</Tabs.Trigger>
							</Tabs.List>
						</Tabs.Root>
						<Tabs.Root value="line">
							<Tabs.List variant="line">
								<Tabs.Trigger value="line">Active line tab</Tabs.Trigger>
								<Tabs.Trigger value="other">Other tab</Tabs.Trigger>
							</Tabs.List>
						</Tabs.Root>
					</div>
				</div>
			</div>

			<div class="border-border bg-card rounded-md border p-5">
				<h2 class="text-sm font-semibold">Forest navigation</h2>
				<nav
					class="bg-sidebar text-sidebar-foreground mt-4 rounded-md p-2"
					aria-label="Specimen navigation"
				>
					<Sidebar.Menu>
						<Sidebar.MenuItem><Sidebar.MenuButton>Overview</Sidebar.MenuButton></Sidebar.MenuItem>
						<Sidebar.MenuItem
							><Sidebar.MenuButton isActive>Analysis</Sidebar.MenuButton></Sidebar.MenuItem
						>
						<Sidebar.MenuItem
							><Sidebar.MenuButton>Report with a deliberately long label</Sidebar.MenuButton
							></Sidebar.MenuItem
						>
					</Sidebar.Menu>
				</nav>
				<Button class="mt-4" variant="outline" onclick={() => (dialogOpen = true)}
					>Open overlay</Button
				>
			</div>
		</section>

		<section class="border-border bg-card rounded-md border p-5">
			<div class="flex items-baseline justify-between gap-4">
				<div>
					<h2 class="text-sm font-semibold">Recent activity</h2>
					<p class="text-muted-foreground mt-1 text-xs">Card surface with tabular times.</p>
				</div>
				<Badge variant="success">3 healthy</Badge>
			</div>
			<div class="mt-4 overflow-hidden rounded-md border border-border">
				<Table.Root>
					<Table.Header
						><Table.Row
							><Table.Head>Activity</Table.Head><Table.Head>Status</Table.Head><Table.Head
								class="text-right">Updated</Table.Head
							></Table.Row
						></Table.Header
					>
					<Table.Body>
						<Table.Row data-state="selected"
							><Table.Cell class="font-medium">Checkout search timeouts</Table.Cell><Table.Cell
								><Badge variant="success">Resolved</Badge></Table.Cell
							><Table.Cell class="text-muted-foreground text-right">14:07</Table.Cell
							></Table.Row
						>
						<Table.Row
							><Table.Cell class="font-medium">Cache pressure after deployment</Table.Cell
							><Table.Cell><Badge variant="warning">Watching</Badge></Table.Cell><Table.Cell
								class="text-muted-foreground text-right">14:12</Table.Cell
							></Table.Row
						>
						<Table.Row
							><Table.Cell class="font-medium"
								>Raw event ID <code class="text-muted-foreground ml-2 font-mono text-xs"
									>a8c92f1</code
								></Table.Cell
							><Table.Cell><Badge variant="neutral">Unknown</Badge></Table.Cell><Table.Cell
								class="text-muted-foreground text-right">14:18</Table.Cell
							></Table.Row
						>
					</Table.Body>
				</Table.Root>
			</div>
		</section>

		<section class="border-border bg-card rounded-md border p-5">
			<h2 class="text-sm font-semibold">Chart palette</h2>
			<div class="mt-4 flex flex-wrap gap-3 text-xs text-muted-foreground">
				<div class="flex items-center gap-2">
					<span class="bg-chart-1 size-3 rounded-full"></span>Series 1
				</div>
				<div class="flex items-center gap-2">
					<span class="bg-chart-2 size-3 rounded-full"></span>Series 2
				</div>
				<div class="flex items-center gap-2">
					<span class="bg-chart-3 size-3 rounded-full"></span>Series 3
				</div>
				<div class="flex items-center gap-2">
					<span class="bg-chart-4 size-3 rounded-full"></span>Series 4
				</div>
				<div class="flex items-center gap-2">
					<span class="bg-chart-5 size-3 rounded-full"></span>Series 5
				</div>
			</div>
		</section>
	</div>
</div>

<Dialog.Root bind:open={dialogOpen}>
	<Dialog.Content>
		<Dialog.Header
			><Dialog.Title>Review evidence</Dialog.Title><Dialog.Description
				>A compact portal surface with preserved dismissal and focus behavior.</Dialog.Description
			></Dialog.Header
		>
		<p class="text-muted-foreground text-sm">This specimen uses illustrative content only.</p>
		<Dialog.Footer
			><Button variant="outline" onclick={() => (dialogOpen = false)}><RiCloseLine />Close</Button
			><Button onclick={() => (dialogOpen = false)}>Continue</Button></Dialog.Footer
		>
	</Dialog.Content>
</Dialog.Root>
