export type NavKind = 'link' | 'push' | 'replace' | 'pop' | 'load';

export type FlowStatus = 'entered' | 'completed' | 'abandoned' | 'failed';

/** Snake-case ingest event. Undefined fields are omitted by JSON.stringify. */
export interface WireEvent {
  event_id: string;
  type: string;
  timestamp: number;
  mono_ms: number;
  page_seq: number;
  anonymous_id: string;
  session_id: string;
  route: string;
  flow_id: string;
  flow_name: string;
  app_version: string;
  app_build: string;
  app_namespace: string;
  release_id: string;
  environment: string;
  element_signature?: string;
  form_key?: string;
  click_id?: string;
  result?: string;
  reason?: string;
  elapsed_ms?: number;
  from_route?: string;
  to_route?: string;
  kind?: string;
  field_keys?: string[];
  error_code?: string;
  name?: string;
  message_slug?: string;
  file?: string;
  function?: string;
  line?: number;
  column?: number;
  truncated?: boolean;
  method?: string;
  status?: string | number;
  path?: string;
  props?: Record<string, string | number | boolean>;
  step?: string;
  attempt_id?: string;
  steps?: string[];
  window_ms?: number;
  important?: boolean;
}
