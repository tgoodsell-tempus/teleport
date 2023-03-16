export type AllowedTargets =
  | 'user'
  | 'role'
  | 'login'
  | 'node'
  | 'mfa_device'
  | 'windows_desktop'
  | 'access_request'
  | 'device';

export type TableData = {
  [key: string]: string;
};

export type LockTarget = {
  label: string;
  value: AllowedTargets;
};

export type SelectedLockTarget = {
  type: AllowedTargets;
  name: string;
};
