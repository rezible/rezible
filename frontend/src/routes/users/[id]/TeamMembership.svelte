<script lang="ts">
  import type { User } from "$lib/api/oapi.gen/types.gen";
  
  export let user: User;
  export let teams: any[] = [];
</script>

<div class="team-membership card">
  <h2>Team Membership</h2>
  
  {#if teams.length === 0}
    <div class="empty-state">
      <p>Not a member of any teams</p>
    </div>
  {:else}
    <ul class="teams-list">
      {#each teams as team}
        <li>
          <a href="/teams/{team.id}" class="team-link">
            <div class="team-icon" style="background-color: {team.color || '#4299E1'}">
              {team.name.substring(0, 2).toUpperCase()}
            </div>
            <div class="team-info">
              <div class="team-name">{team.name}</div>
              {#if team.role}
                <div class="team-role">{team.role}</div>
              {/if}
            </div>
          </a>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .team-membership {
    margin-top: 1.5rem;
  }
  
  h2 {
    font-size: 1.25rem;
    font-weight: 600;
    margin-bottom: 1rem;
    color: var(--color-text-primary);
  }
  
  .empty-state {
    padding: 2rem;
    text-align: center;
    color: var(--color-text-secondary);
    background-color: var(--color-bg-subtle);
    border-radius: 6px;
  }
  
  .teams-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 1rem;
  }
  
  .team-link {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    border-radius: 6px;
    background-color: var(--color-bg-subtle);
    text-decoration: none;
    color: inherit;
    transition: background-color 0.2s;
  }
  
  .team-link:hover {
    background-color: var(--color-bg-hover);
  }
  
  .team-icon {
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: 600;
    font-size: 0.875rem;
  }
  
  .team-name {
    font-weight: 500;
  }
  
  .team-role {
    font-size: 0.75rem;
    color: var(--color-text-secondary);
  }
</style>
