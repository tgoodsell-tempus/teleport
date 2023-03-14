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

import { Box } from 'design';
import Select from 'shared/components/Select';
import Table, { Cell } from 'design/DataTable';

import cfg from 'teleport/config';
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

function filterTargetData(targetType: AllowedTargets, data: any): TableData[] {
  if (targetType === 'user') {
    return data.map(u => ({
      name: u.name,
      roles: u.roles.join(', '),
    }));
  }
}

const useGetTargetData = (targetType: AllowedTargets, clusterId: string) => {
  const [targetData, setTargetData] = useState<TableData[]>();
  const { userService } = useTeleport();

  useEffect(() => {
    if (targetType == 'user') {
      userService.fetchUsers().then(users => {
        const filteredData = filterTargetData('user', users);
        setTargetData(filteredData);
      });
    }
  }, [targetType, userService]);

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
  const [selectedTargetType, setSelectedTargetType] = useState<LockTarget>();
  const { clusterId } = useStickyClusterId();
  const targetData = useGetTargetData(selectedTargetType?.value, clusterId);

  return (
    <FeatureBox>
      <FeatureHeader>
        <FeatureHeaderTitle>
          Session & Identity Locks / Create New Lock
        </FeatureHeaderTitle>
      </FeatureHeader>
      <Box width="150px" mb={4} data-testid="resource-selector">
        <Select
          value={selectedTargetType}
          options={lockTargets}
          onChange={(o: LockTarget) => setSelectedTargetType(o)}
        />
      </Box>
      <TargetList data={targetData} />
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
};

export function TargetList({ data }: TargetListProps) {
  if (!data) data = [];

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
