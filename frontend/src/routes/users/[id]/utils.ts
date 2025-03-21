import type { User } from "$lib/api/oapi.gen/types.gen";

// Mock data generators for demonstration purposes
export function generateMockOncallShifts(userId: string) {
  const now = new Date();
  const shifts = [];
  
  // Past shifts
  for (let i = 1; i <= 5; i++) {
    const startDate = new Date(now);
    startDate.setDate(now.getDate() - (i * 7));
    
    const endDate = new Date(startDate);
    endDate.setDate(startDate.getDate() + 7);
    
    shifts.push({
      id: `shift-${i}`,
      userId,
      rosterId: `roster-${i % 3 + 1}`,
      roster: { name: `Team ${i % 3 + 1} Rotation` },
      startAt: startDate.toISOString(),
      endAt: endDate.toISOString()
    });
  }
  
  // Current shift (50% chance)
  if (Math.random() > 0.5) {
    const startDate = new Date(now);
    startDate.setDate(now.getDate() - 3);
    
    const endDate = new Date(startDate);
    endDate.setDate(startDate.getDate() + 7);
    
    shifts.push({
      id: 'current-shift',
      userId,
      rosterId: 'roster-current',
      roster: { name: 'Primary On-Call' },
      startAt: startDate.toISOString(),
      endAt: endDate.toISOString()
    });
  }
  
  // Future shifts
  for (let i = 1; i <= 3; i++) {
    const startDate = new Date(now);
    startDate.setDate(now.getDate() + (i * 14));
    
    const endDate = new Date(startDate);
    endDate.setDate(startDate.getDate() + 7);
    
    shifts.push({
      id: `future-shift-${i}`,
      userId,
      rosterId: `roster-future-${i}`,
      roster: { name: `Team ${i % 3 + 1} Rotation` },
      startAt: startDate.toISOString(),
      endAt: endDate.toISOString()
    });
  }
  
  return shifts;
}

export function generateMockIncidents(userId: string) {
  const severities = ['critical', 'high', 'medium', 'low'];
  const statuses = ['active', 'resolved', 'resolved', 'resolved']; // 75% resolved
  const roles = ['incident_lead', 'communications_lead', 'technical_lead', 'participant'];
  const now = new Date();
  
  return Array(10).fill(0).map((_, i) => {
    const createdAt = new Date(now);
    createdAt.setDate(now.getDate() - Math.floor(Math.random() * 90)); // Random date in last 90 days
    
    const severity = severities[Math.floor(Math.random() * severities.length)];
    const status = statuses[Math.floor(Math.random() * statuses.length)];
    const role = roles[Math.floor(Math.random() * roles.length)];
    
    return {
      id: `incident-${i}`,
      slug: `incident-${i}`,
      title: `Incident ${i}: ${severity.charAt(0).toUpperCase() + severity.slice(1)} Service Disruption`,
      severity,
      status,
      createdAt: createdAt.toISOString(),
      roles: [
        {
          userId,
          role
        }
      ]
    };
  });
}

export function generateMockTeams(userId: string) {
  const teamColors = ['#4299E1', '#48BB78', '#ED8936', '#9F7AEA', '#F56565'];
  const roles = ['Member', 'Lead', 'Manager', 'Observer'];
  
  return Array(Math.floor(Math.random() * 4) + 1).fill(0).map((_, i) => {
    return {
      id: `team-${i}`,
      name: `Team ${['Alpha', 'Beta', 'Platform', 'Infrastructure', 'Frontend', 'Backend'][i % 6]}`,
      color: teamColors[i % teamColors.length],
      role: roles[Math.floor(Math.random() * roles.length)]
    };
  });
}
