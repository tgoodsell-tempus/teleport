import { LabelDescription } from 'design/DataTable/types';

export type Lock = {
  name: string;
  message: string;
  expires: string;
  createdAt: string;
  createdBy: string;
  targets: {
    user?: string;
    role?: string;
    login?: string;
    node?: string;
    mfa_device?: string;
    windows_desktop?: string;
    access_request?: string;
    device?: string;
  };
};

export type LockForTable = {
  name: string;
  message: string;
  expires: string;
  createdAt: string;
  createdBy: string;
  targets: LabelDescription[];
};

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

export type OnAdd = (name: string) => void;

export type TargetListProps = {
  data: TableData[];
  onAdd: OnAdd;
  selectedTarget: AllowedTargets;
};

export type CreateLockData = {
  targets: { [K in AllowedTargets]?: string };
  message?: string;
  ttl?: string;
};
