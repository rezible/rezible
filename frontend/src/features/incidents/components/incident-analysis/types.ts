// Core event type
interface TimelineEvent {
	id: string;
	incidentId: string;
	timestamp: Date | null;  // null for "unknown time"
	type: 'observation' | 'action' | 'decision' | 'context';
	title: string;
	description: string;
	createdAt: Date;
	updatedAt: Date;
	createdBy: string;
	sequence: number;  // for ordering events with same timestamp
	isDraft: boolean;
  }
  
  // Context details
  interface EventContext {
	id: string;
	eventId: string;
	systemState?: string;
	decisionOptions?: string[];
	decisionRationale?: string;
	involvedPersonnel?: string[];
  }
  
  // Contributing factors
  interface ContributingFactor {
	id: string;
	eventId: string;
	type: string;  // from controlled vocabulary
	description: string;
  }
  
  // Evidence links
  interface EventEvidence {
	id: string;
	eventId: string;
	type: 'log' | 'metric' | 'chat' | 'ticket' | 'other';
	url: string;
	title: string;
	description?: string;
  }
  
  // Component relationships
  interface EventComponent {
	id: string;
	eventId: string;
	componentId: string;
	relationship: 'primary' | 'affected' | 'contributing';
  }