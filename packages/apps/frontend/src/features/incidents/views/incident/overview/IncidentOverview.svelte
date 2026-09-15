<script lang="ts">
	import { resolve } from "$app/paths";
	import * as Card from "$components/ui/card";
	import * as Field from "$components/ui/field";
	import { Separator } from "$components/ui/separator";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { Input } from "$components/ui/input";
	import { Textarea } from "$components/ui/textarea";
	import { initIncidentOverviewController } from "./controller.svelte";

	const statusLabel = (status: string) => status.replace("_", " ").replace(/^./, (c) => c.toUpperCase());

	const overview = initIncidentOverviewController();
	const incAttrs = $derived(overview.incident?.attributes);
	const start = $derived(incAttrs?.openedAt ? new Date(incAttrs.openedAt).toLocaleString() : undefined);

	const retroAttrs = $derived(overview.retrospective?.attributes);
</script>

<div class="min-h-0 flex-1 overflow-y-auto">
	<div class="mx-auto grid w-full max-w-[1240px] gap-8 p-6 lg:grid-cols-[minmax(0,1fr)_320px]">
		<main class="min-w-0">
			<div class="mb-6 flex items-start justify-between gap-4">
				<div class="min-w-0 flex-1">
					{#if overview.editing}
						<Field.FieldGroup>
							<Field.Field>
								<Field.FieldLabel for="incident-title">Title</Field.FieldLabel>
								<Input id="incident-title" bind:value={overview.title} />
							</Field.Field>
							<Field.Field>
								<Field.FieldLabel for="incident-summary">Summary</Field.FieldLabel>
								<Textarea id="incident-summary" bind:value={overview.summary} rows={3} />
							</Field.Field>
						</Field.FieldGroup>
					{:else}
						<h1 class="text-3xl font-semibold tracking-tight">
							{incAttrs?.title ?? "Incident"}
						</h1>
						<p class="mt-2 text-lg text-muted-foreground">{incAttrs?.summary}</p>
					{/if}
				</div>
				{#if overview.editing}
					<div class="flex w-full shrink-0 flex-col gap-2 sm:w-auto sm:flex-row">
						<Button
							variant="outline"
							disabled={overview.saving}
							onclick={() => (overview.editing = false)}
						>Cancel</Button>
						<Button 
							disabled={overview.saving} 
							onclick={overview.save}
						>{overview.saving ? "Saving…" : "Save"}</Button>
					</div>
				{:else}
					<Button variant="ghost" onclick={overview.beginEdit}>Edit</Button>
				{/if}
			</div>
			
			{#if overview.error}
				<p class="mb-4 text-sm text-destructive" role="alert">{overview.error}</p>
			{/if}

			<Card.Root>
				<Card.Header><Card.Title>Incident details</Card.Title></Card.Header>
				<Card.Content class="grid gap-5">
					<div class="grid gap-5 sm:grid-cols-3">
						<div>
							<p class="text-sm text-muted-foreground">Severity</p>
							<Badge class="mt-2">{incAttrs?.severity.attributes.name ?? "—"}</Badge>
						</div>
						<div>
							<p class="text-sm text-muted-foreground">Status</p>
							<p class="mt-2 font-medium text-primary">
								{statusLabel(incAttrs?.currentStatus ?? "")}
							</p>
						</div>
					</div>
					<Separator />
					<div class="grid gap-5 sm:grid-cols-3">
						<div>
							<p class="text-sm text-muted-foreground">Started</p>
							<p class="mt-2 font-medium">{start ?? "—"}</p>
						</div>
					</div>
				</Card.Content>
			</Card.Root>
		</main>
		<aside class="min-w-0">
			<Card.Root>
				<Card.Header><Card.Title>Review status</Card.Title></Card.Header>
				<Card.Content>
					{#if incAttrs && retroAttrs}
						{@const reportHref = resolve("/incidents/[slug]/[[view=incidentView]]", {
							slug: incAttrs.slug,
							view: "report",
						})}
						<p class="font-medium">{statusLabel(retroAttrs.state)}</p>
						<a href={reportHref} class="mt-5 inline-block font-medium text-primary underline">
							Open report
						</a>
					{:else}
						<p class="text-sm text-muted-foreground">
							No retrospective is associated with this incident.
						</p>
					{/if}
				</Card.Content>
			</Card.Root>
		</aside>
	</div>
</div>
