import { useEffect, useState } from 'react';

import useTeleport from 'teleport/useTeleport';

import type { AllowedTargets, LockTarget, TableData } from './types';

export const lockTargets: LockTarget[] = [
  { label: 'User', value: 'user' },
  { label: 'Role', value: 'role' },
  { label: 'Login', value: 'login' },
  { label: 'Node', value: 'node' },
  { label: 'MFA Device', value: 'mfa_device' },
  { label: 'Windows Desktop', value: 'windows_desktop' },
  { label: 'Device', value: 'device' },
];

export type UseGetTargetData = (
  targetType: AllowedTargets,
  clusterId: string,
  additionalTargets?: AdditionalTargets
) => TableData[];

export type AdditionalTargets = {
  [key: string]: {
    fetch: (options: any) => Promise<any>;
    handler: (setter: (data: TableData[]) => void, data: any) => void;
    options: any;
  };
};

export const useGetTargetData: UseGetTargetData = (
  targetType,
  clusterId,
  additionalTargets
) => {
  const [targetData, setTargetData] = useState<TableData[]>();
  const {
    desktopService,
    mfaService,
    nodeService,
    resourceService,
    userService,
  } = useTeleport();

  useEffect(() => {
    const targetDataFilters = {
      user: {
        fetch: userService.fetchUsers,
        handler: (setter, users) => {
          const filteredData = users.map(u => ({
            name: u.name,
            roles: u.roles.join(', '),
          }));
          setter(filteredData);
        },
      },
      role: {
        fetch: resourceService.fetchRoles,
        handler: (setter, roles) => {
          const filteredData = roles.map(r => ({
            name: r.name,
          }));
          setter(filteredData);
        },
      },
      node: {
        fetch: nodeService.fetchNodes,
        handler: (setter, nodes) => {
          const filteredData = nodes.agents.map(n => ({
            hostname: n.hostname,
            addr: n.addr,
            labels: n.labels.join(', '),
          }));
          setter(filteredData);
        },
        options: [
          clusterId,
          {
            limit: 10,
          },
        ],
      },
      mfa_device: {
        fetch: mfaService.fetchDevices,
        handler: (setter, mfas) => {
          const filteredData = mfas.map(m => ({
            name: m.name,
            id: m.id,
            description: m.description,
            lastUsed: m.lastUsedDate.toUTCString(),
          }));
          setter(filteredData);
        },
      },
      windows_desktop: {
        fetch: desktopService.fetchDesktops,
        handler: (setter, desktops) => {
          const filteredData = desktops.agents.map(d => ({
            name: d.name,
            addr: d.addr,
            labels: d.labels.map(l => `${l.name}:${l.value}`).join(', '),
          }));
          setter(filteredData);
        },
        options: [clusterId, { limit: 10 }],
      },
    };

    let action =
      targetDataFilters[targetType] || additionalTargets?.[targetType];
    if (!action) {
      console.log(`unknown target type ${targetType}`);
      setTargetData([]);
      return;
    }

    action.fetch
      .apply(null, action.options)
      .then(action.handler.bind(null, setTargetData));
  }, [
    additionalTargets,
    clusterId,
    desktopService.fetchDesktops,
    mfaService.fetchDevices,
    nodeService.fetchNodes,
    resourceService.fetchRoles,
    targetType,
    userService.fetchUsers,
  ]);

  return targetData;
};
