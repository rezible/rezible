<script lang="ts">
  import type { User } from "$lib/api/oapi.gen/types.gen";
  
  export let user: User;
  export let oncallShifts: any[] = [];
  
  // Calculate stats
  const totalShifts = oncallShifts.length;
  const currentlyOncall = oncallShifts.some(shift => {
    const now = new Date();
    const start = new Date(shift.startAt);
    const end = new Date(shift.endAt);
    return start <= now && end >= now;
  });
  
  const hoursOncall = oncallShifts.reduce((total, shift) => {
    const start = new Date(shift.startAt);
    const end = new Date(shift.endAt);
    const hours = (end.getTime() - start.getTime()) / (1000 * 60 * 60);
    return total + hours;
  }, 0);
  
  const upcomingShifts = oncallShifts
    .filter(shift => new Date(shift.startAt) > new Date())
    .sort((a, b) => new Date(a.startAt).getTime() - new Date(b.startAt).getTime())
    .slice(0, 3);
</script>

<div class="oncall-stats card">
  <h2>Oncall Statistics</h2>
  
  <div class="stats-grid">
    <div class="stat-item">
      <div class="stat-value">{totalShifts}</div>
      <div class="stat-label">Total Shifts</div>
    </div>
    
    <div class="stat-item">
      <div class="stat-value">{Math.round(hoursOncall)}</div>
      <div class="stat-label">Hours Oncall</div>
    </div>
    
    <div class="stat-item">
      <div class="stat-value status-indicator {currentlyOncall ? 'active' : 'inactive'}">
        {currentlyOncall ? 'Yes' : 'No'}
      </div>
      <div class="stat-label">Currently Oncall</div>
    </div>
  </div>
  
  {#if upcomingShifts.length > 0}
    <div class="upcoming-shifts">
      <h3>Upcoming Shifts</h3>
      <ul>
        {#each upcomingShifts as shift}
          <li>
            <div class="shift-roster">{shift.roster?.name || 'Unknown Roster'}</div>
            <div class="shift-time">
              {new Date(shift.startAt).toLocaleDateString()} - 
              {new Date(shift.endAt).toLocaleDateString()}
            </div>
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</div>

<style>
  .oncall-stats {
    margin-top: 1.5rem;
  }
  
  h2 {
    font-size: 1.25rem;
    font-weight: 600;
    margin-bottom: 1rem;
    color: var(--color-text-primary);
  }
  
  h3 {
    font-size: 1rem;
    font-weight: 600;
    margin: 1.5rem 0 0.75rem 0;
    color: var(--color-text-primary);
  }
  
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
    gap: 1rem;
  }
  
  .stat-item {
    background-color: var(--color-bg-subtle);
    padding: 1rem;
    border-radius: 6px;
    text-align: center;
  }
  
  .stat-value {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-text-primary);
    margin-bottom: 0.25rem;
  }
  
  .stat-label {
    font-size: 0.875rem;
    color: var(--color-text-secondary);
  }
  
  .status-indicator {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 1rem;
  }
  
  .status-indicator.active {
    background-color: var(--color-success-bg);
    color: var(--color-success-text);
  }
  
  .status-indicator.inactive {
    background-color: var(--color-neutral-bg);
    color: var(--color-neutral-text);
  }
  
  .upcoming-shifts ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  
  .upcoming-shifts li {
    padding: 0.75rem;
    border-radius: 6px;
    background-color: var(--color-bg-subtle);
    margin-bottom: 0.5rem;
  }
  
  .shift-roster {
    font-weight: 600;
    margin-bottom: 0.25rem;
  }
  
  .shift-time {
    font-size: 0.875rem;
    color: var(--color-text-secondary);
  }
</style>
