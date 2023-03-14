/*
Copyright 2023 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import React, { useEffect, useState } from 'react';
import styled from 'styled-components';

import { Box, ButtonPrimary, Flex, Text, Input } from 'design';
import Select from 'shared/components/Select';
import Table, { Cell } from 'design/DataTable';

import useTeleport from 'teleport/useTeleport';
import useStickyClusterId from 'teleport/useStickyClusterId';

import {
  FeatureBox,
  FeatureHeader,
  FeatureHeaderTitle,
} from 'teleport/components/Layout';

type TableData = {
  [key: string]: string;
};

const useGetTargetData = (targetType: AllowedTargets, clusterId: string) => {
  const [targetData, setTargetData] = useState<TableData[]>();
  const { userService, resourceService, nodeService, mfaService } =
    useTeleport();

  useEffect(() => {
    switch (targetType) {
      case 'user':
        userService.fetchUsers().then(users => {
          const filteredData = users.map(u => ({
            name: u.name,
            roles: u.roles.join(', '),
          }));
          setTargetData(filteredData);
        });
        break;
      case 'role':
        resourceService.fetchRoles().then(roles => {
          const filteredData = roles.map(r => ({
            name: r.name,
          }));
          setTargetData(filteredData);
        });
        break;
      case 'node':
        nodeService
          .fetchNodes(clusterId, {
            limit: 10,
          })
          .then(nodes => {
            const filteredData = nodes.agents.map(n => ({
              hostname: n.hostname,
              addr: n.addr,
              labels: n.labels.join(', '),
            }));
            setTargetData(filteredData);
          });
        break;
      case 'mfa_device':
        mfaService.fetchDevices().then(mfas => {
          const filteredData = mfas.map(m => ({
            name: m.name,
            id: m.id,
            description: m.description,
            lastUsed: m.lastUsedDate.toUTCString(),
          }));
          setTargetData(filteredData);
        });
        break;
    }
  }, [
    clusterId,
    mfaService,
    nodeService,
    resourceService,
    targetType,
    userService,
  ]);

  return targetData;
};

type AllowedTargets =
  | 'user'
  | 'role'
  | 'login'
  | 'node'
  | 'mfa_device'
  | 'windows_desktop'
  | 'access_request'
  | 'device';

type LockTarget = {
  label: string;
  value: AllowedTargets;
};

const lockTargets: LockTarget[] = [
  { label: 'User', value: 'user' },
  { label: 'Role', value: 'role' },
  { label: 'Login', value: 'login' },
  { label: 'Node', value: 'node' },
  { label: 'MFA Device', value: 'mfa_device' },
  { label: 'Windows Desktop', value: 'windows_desktop' },
  { label: 'Access Request', value: 'access_request' },
  { label: 'Device', value: 'device' },
];

export default function NewLock() {
  const [selectedTargetType, setSelectedTargetType] = useState<LockTarget>({
    label: 'User',
    value: 'user',
  });
  const { clusterId } = useStickyClusterId();
  const targetData = useGetTargetData(selectedTargetType?.value, clusterId);

  return (
    <FeatureBox>
      <FeatureHeader>
        <FeatureHeaderTitle>
          Session & Identity Locks / Create New Lock
        </FeatureHeaderTitle>
      </FeatureHeader>
      <Flex justifyContent="space-between">
        <Box width="150px" mb={4} data-testid="resource-selector">
          <Select
            value={selectedTargetType}
            options={lockTargets}
            onChange={(o: LockTarget) => setSelectedTargetType(o)}
          />
        </Box>
        <QuickAdd targetType={selectedTargetType.label} />
      </Flex>
      <TargetList data={targetData} selectedTarget={selectedTargetType.value} />
    </FeatureBox>
  );
}

const StyledTable = styled(Table)`
  & > tbody > tr > td {
    vertical-align: middle;
  }
` as typeof Table;

type TargetListProps = {
  data: TableData[];
  selectedTarget: AllowedTargets;
};

function TargetList({ data, selectedTarget }: TargetListProps) {
  if (!data) data = [];

  if (selectedTarget === 'device') {
    return <Box>Not Implemented</Box>;
  }

  const columns = data.length
    ? Object.keys(data[0]).map(c => ({
        key: c,
        headerText: c,
        isSortable: true,
      }))
    : [];
  return (
    <StyledTable data={data} columns={columns} emptyText="No Targets Found" />
  );
}

function QuickAdd({ targetType }: { targetType: string }) {
  return (
    <Flex
      justifyContent="flex-end"
      alignItems="baseline"
      css={{ columnGap: '20px' }}
      mb={4}
    >
      <Input placeholder={`Quick add ${targetType}`} width={500} />
      <ButtonPrimary>Add</ButtonPrimary>
    </Flex>
  );
}
