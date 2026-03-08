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

export interface InterfaceTaskCandidate {
  name: string;
}

export interface InterfaceParseResult {
  interface_path: string;
  base_dir: string;
  name: string;
  version: string;
  controller_candidates: InterfaceControllerCandidate[];
  resource_candidates: InterfaceResourceCandidate[];
  task_candidates: InterfaceTaskCandidate[];
}
