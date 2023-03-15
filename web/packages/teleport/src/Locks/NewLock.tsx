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

import React, { useState } from 'react';
import styled from 'styled-components';

import { Box, ButtonPrimary, Flex, Input } from 'design';
import Select from 'shared/components/Select';
import Table, { Cell } from 'design/DataTable';

import useStickyClusterId from 'teleport/useStickyClusterId';

import {
  FeatureBox,
  FeatureHeader,
  FeatureHeaderTitle,
} from 'teleport/components/Layout';

import { lockTargets, useGetTargetData } from './useGetTargetData';

import type { AdditionalTargets } from './useGetTargetData';
import type { AllowedTargets, LockTarget, TableData } from './types';

// This is split out like this to allow the router to call 'NewLock'
// but also allow E to use 'NewLockContent' separately.
export default function NewLock() {
  return <NewLockContent />;
}

export function NewLockContent({
  additionalTargets,
}: {
  additionalTargets?: AdditionalTargets;
}) {
  const [selectedTargetType, setSelectedTargetType] = useState<LockTarget>({
    label: 'User',
    value: 'user',
  });
  const { clusterId } = useStickyClusterId();

  const targetData = useGetTargetData(
    selectedTargetType?.value,
    clusterId,
    additionalTargets
  );

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
    return <Box>Listing Devices not implemented.</Box>;
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
