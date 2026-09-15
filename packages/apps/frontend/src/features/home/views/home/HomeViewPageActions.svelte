<script>
	import { Label } from "$components/ui/label";
	import { Input } from "$components/ui/input";
    import RiSearch2Line from "remixicon-svelte/icons/search-2-line";
</script>

<div class="flex items-center relative">
    <Label for="search" class="sr-only">Search</Label>
    <Input id="search" placeholder="Search..." class="ps-8" />
    <RiSearch2Line class="pointer-events-none absolute start-2 top-1/2 size-4 -translate-y-1/2 opacity-50 select-none" />
    <!-- <Badge class="absolute end-2 top-1/2 -translate-y-1/2 opacity-50">CMD-K</Badge> -->
</div>