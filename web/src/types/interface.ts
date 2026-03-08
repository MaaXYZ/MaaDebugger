export interface InterfaceResolvedPath {
  source: string;
  path: string;
  exists: boolean;
}

export interface InterfaceControllerCandidate {
  name: string;
  type: string;
  class_regex?: string;
  window_regex?: string;
  uuid?: string;
  attach_resource_paths?: string[];
}

export interface InterfaceResourceCandidate {
  name: string;
  label?: string;
  resolved_paths: InterfaceResolvedPath[];
}

export interface InterfaceTaskOptionCase {
  name: string;
  label?: string;
  description?: string;
  pipeline_override_keys?: string[];
}

export interface InterfaceTaskOptionDefinition {
  name: string;
  type?: string;
  label?: string;
  description?: string;
  default_case?: string;
  cases?: InterfaceTaskOptionCase[];
  source?: string;
  resolved_from?: string;
}

export interface InterfaceTaskCandidate {
  name: string;
  label?: string;
  entry?: string;
  description?: string;
  controllers?: string[];
  resources?: string[];
  options?: string[];
  option_defs?: InterfaceTaskOptionDefinition[];
  source?: string;
  source_interface?: string;
}

export interface InterfaceParseResult {
  interface_path: string;
  base_dir: string;
  name: string;
  version: string;
  imports?: InterfaceResolvedPath[];
  controller_candidates: InterfaceControllerCandidate[];
  resource_candidates: InterfaceResourceCandidate[];
  task_candidates: InterfaceTaskCandidate[];
}
