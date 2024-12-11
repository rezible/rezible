<script lang="ts">
    import { mdiChartBar, mdiChevronRight } from "@mdi/js";
    import { Button, Header, ListItem } from "svelte-ux";
</script>

<div class="flex flex-col gap-2 rounded-lg p-2 border">
    <Header title="Pinned Items" class="border-b" />
    
    <div class="flex flex-col gap-2 overflow-y-auto">
        <a href="/reports">
            <ListItem 
                title="Search Team Oncall Health Report"
                classes={{ root: 'hover:bg-primary-900/80' }} 
                icon={mdiChartBar}
            >
                <div slot="actions">
                    <Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
                </div>
            </ListItem>
        </a>
    </div>
</div>