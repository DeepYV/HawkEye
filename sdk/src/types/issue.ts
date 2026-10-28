export type IssueType =
  | 'repeated_click'
  | 'dead_click'
  | 'validation_loop'
  | 'failed_action'
  | 'flow_abandoned';

export type Confidence = 'high' | 'medium' | 'low';

export interface IssueEvidence {
  confidenceRule: string;
  maxBurst?: number;
  truncated?: boolean;
  formKey?: string;
  fieldKeys?: string[];
  errorCode?: string;
  kind?: 'submit' | 'http' | 'js';
  method?: string;
  status?: number;
  path?: string;
  name?: string;
  file?: string;
  function?: string;
  step?: string;
  endReason?: 'window' | 'host_abandoned' | 'host_failed' | 'restart';
  attemptsReached?: number;
  sessionCount?: number;
  firstSeenAppVersion?: string;
  lastSeenAppVersion?: string;
  appBuild?: string;
  appNamespace?: string;
  releaseId?: string;
}

export interface Issue {
  id: string;
  type: IssueType;
  route: string;
  elementSignature: string;
  flowName: string;
  users: number;
  occurrences: number;
  impactScore: number;
  confidence: Confidence;
  observed: string;
  explanation: string;
  suggestion: string;
  firstSeen: number;
  lastSeen: number;
  release: string;
  important: boolean;
  evidence: IssueEvidence;
}

export interface FlowStepSummary {
  step: string;
  reached: number;
  abandoned: number;
}

export interface FlowSummary {
  name: string;
  important: boolean;
  attempts: number;
  completed: number;
  abandoned: number;
  open: number;
  steps: FlowStepSummary[];
}
