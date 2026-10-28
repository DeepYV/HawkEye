export type FlowDefinition = {
  name: string;
  steps: string[];
  windowMs?: number;
  important?: boolean;
};

export type InitOptions = {
  apiKey: string;
  ingestionUrl: string;
  appVersion?: string;
  appBuild?: string;
  appNamespace?: string;
  releaseId?: string;
  environment?: string;
  /** Full flow registry. `important: true` is the rank weight. Omitted entries are unknown flows. */
  importantFlows?: FlowDefinition[];
  /** Names `track` may send. Any other name is dropped. */
  events?: string[];
};

export type FormResult = 'success' | 'validation_error' | 'submit_error';

export type FormOutcomeInput = {
  formKey: string;
  result: FormResult;
  fieldKeys?: string[];
  errorCode?: string;
};

export type FlowHandle = {
  start(): void;
  step(step: string): void;
  complete(): void;
  fail(reason?: string): void;
  abandon(reason?: string): void;
};
