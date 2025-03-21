<script lang="ts">
  import type { User } from "$lib/api/oapi.gen/types.gen";
  
  export let user: User;
  export let incidents: any[] = [];
  
  // Calculate stats
  const totalIncidents = incidents.length;
  const resolvedIncidents = incidents.filter(inc => inc.status === 'resolved').length;
  const leadIncidents = incidents.filter(inc => inc.roles?.some(r => 
    r.userId === user.id && r.role === 'incident_lead'
  )).length;
  
  // Get recent incidents (last 5)
  const recentIncidents = [...incidents]
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .slice(0, 5);
    
  // Calculate severity distribution
  const severityCount = incidents.reduce((acc, inc) => {
    const sev = inc.severity || 'unknown';
    acc[sev] = (acc[sev] || 0) + 1;
    return acc;
  }, {});
  
  const getSeverityClass = (severity) => {
    const classes = {
      critical: 'severity-critical',
      high: 'severity-high',
      medium: 'severity-medium',
      low: 'severity-low',
      unknown: 'severity-unknown'
    };
    return classes[severity] || classes.unknown;
  };
</script>

<div class="incident-participation card">
  <h2>Incident Participation</h2>
  
  <div class="stats-grid">
    <div class="stat-item">
      <div class="stat-value">{totalIncidents}</div>
      <div class="stat-label">Total Incidents</div>
    </div>
    
    <div class="stat-item">
      <div class="stat-value">{leadIncidents}</div>
      <div class="stat-label">Led as IC</div>
    </div>
    
    <div class="stat-item">
      <div class="stat-value">{Math.round((resolvedIncidents / totalIncidents) * 100) || 0}%</div>
      <div class="stat-label">Resolution Rate</div>
    </div>
  </div>
  
  {#if Object.keys(severityCount).length > 0}
    <div class="severity-distribution">
      <h3>Severity Distribution</h3>
      <div class="severity-bars">
        {#each Object.entries(severityCount) as [severity, count]}
          <div class="severity-bar-container">
            <div class="severity-label">{severity}</div>
            <div class="severity-bar-wrapper">
              <div 
                class="severity-bar {getSeverityClass(severity)}" 
                style="width: {(count / totalIncidents) * 100}%"
              >
                <span class="severity-count">{count}</span>
              </div>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}
  
  {#if recentIncidents.length > 0}
    <div class="recent-incidents">
      <h3>Recent Incidents</h3>
      <ul>
        {#each recentIncidents as incident}
          <li>
            <a href="/incidents/{incident.slug}" class="incident-link">
              <div class="incident-title">
                <span class="severity-dot {getSeverityClass(incident.severity)}"></span>
                {incident.title}
              </div>
              <div class="incident-meta">
                {new Date(incident.createdAt).toLocaleDateString()}
                {#if incident.roles?.some(r => r.userId === user.id)}
                  • {incident.roles.find(r => r.userId === user.id).role.replace('_', ' ')}
                {/if}
              </div>
            </a>
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</div>

<style>
  .incident-participation {
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
  
  .severity-distribution {
    margin-top: 1.5rem;
  }
  
  .severity-bars {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .severity-bar-container {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  
  .severity-label {
    width: 80px;
    font-size: 0.875rem;
    text-transform: capitalize;
  }
  
  .severity-bar-wrapper {
    flex: 1;
    background-color: var(--color-bg-subtle);
    border-radius: 4px;
    overflow: hidden;
    height: 1.5rem;
  }
  
  .severity-bar {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding: 0 0.5rem;
    min-width: 2rem;
    transition: width 0.3s ease;
  }
  
  .severity-count {
    color: white;
    font-size: 0.75rem;
    font-weight: 600;
  }
  
  .severity-critical {
    background-color: var(--color-critical);
  }
  
  .severity-high {
    background-color: var(--color-high);
  }
  
  .severity-medium {
    background-color: var(--color-medium);
  }
  
  .severity-low {
    background-color: var(--color-low);
  }
  
  .severity-unknown {
    background-color: var(--color-neutral);
  }
  
  .recent-incidents ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  
  .recent-incidents li {
    border-bottom: 1px solid var(--color-border);
    padding: 0.75rem 0;
  }
  
  .recent-incidents li:last-child {
    border-bottom: none;
  }
  
  .incident-link {
    display: block;
    text-decoration: none;
    color: inherit;
  }
  
  .incident-link:hover .incident-title {
    color: var(--color-primary);
  }
  
  .incident-title {
    font-weight: 500;
    margin-bottom: 0.25rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  
  .incident-meta {
    font-size: 0.875rem;
    color: var(--color-text-secondary);
  }
  
  .severity-dot {
    width: 0.75rem;
    height: 0.75rem;
    border-radius: 50%;
    display: inline-block;
  }
</style>
