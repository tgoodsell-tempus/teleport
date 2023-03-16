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

import { Box, ButtonPrimary, Flex, Input, Text } from 'design';
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
import type {
  AllowedTargets,
  LockTarget,
  SelectedLockTarget,
  TableData,
} from './types';
import type { TableColumn } from 'design/DataTable/types';

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
  const { clusterId } = useStickyClusterId();
  const [selectedTargetType, setSelectedTargetType] = useState<LockTarget>({
    label: 'User',
    value: 'user',
  });
  const [selectedLockTargets, setSelectedLockTargets] = useState<
    SelectedLockTarget[]
  >([]);

  const targetData = useGetTargetData(
    selectedTargetType?.value,
    clusterId,
    additionalTargets
  );

  function onAdd(name) {
    selectedLockTargets.push({
      type: selectedTargetType.value,
      name,
    });
    setSelectedLockTargets([...selectedLockTargets]);
  }

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
      <TargetList
        data={targetData}
        onAdd={onAdd}
        selectedTarget={selectedTargetType.value}
      />
      <Flex
        data-testid="selected-locks"
        alignItems="center"
        justifyContent="space-between"
        borderRadius={3}
        p={3}
        mt={4}
        css={`
          background: ${({ theme }) => theme.colors.primary.main};
        `}
      >
        {selectedLockTargets.length > 0 ? (
          <StyledTable
            data={selectedLockTargets}
            columns={[
              {
                key: 'type',
                headerText: 'Type',
                isSortable: false,
              },
              {
                key: 'name',
                headerText: 'Name',
                isSortable: false,
              },
            ]}
            emptyText="No Targets Found"
          />
        ) : (
          <Box>
            <Text>Add lock targets to create lock.</Text>
          </Box>
        )}
      </Flex>
      <Flex justifyContent="flex-end" mt={4}>
        <ButtonPrimary width="182px" onClick={() => {}} disabled={false}>
          Lock targets
        </ButtonPrimary>
      </Flex>
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
  onAdd: (name: string) => void;
  selectedTarget: AllowedTargets;
};

function TargetList({ data, selectedTarget, onAdd }: TargetListProps) {
  if (!data) data = [];

  if (selectedTarget === 'device') {
    return <Box>Listing Devices not implemented.</Box>;
  }

  if (selectedTarget === 'login') {
    return <Box>Unable to list logins, use quick add box.</Box>;
  }

  const columns: TableColumn<any>[] = data.length
    ? Object.keys(data[0]).map(c => ({
        key: c,
        headerText: c,
        isSortable: true,
      }))
    : [];

  if (columns.length) {
    columns.push({
      altKey: 'add-btn',
      render: ({ name }) => <AddCell onAdd={onAdd.bind(null, name)} />,
    });
  }
  return (
    <StyledTable data={data} columns={columns} emptyText="No Targets Found" />
  );
}

function QuickAdd({ targetType }: { targetType: string }) {
  return (
    <Flex
      justifyContent="flex-end"
      alignItems="center"
      css={{ columnGap: '20px' }}
      mb={4}
    >
      <Input placeholder={`Quick add ${targetType}`} width={500} />
      <ButtonPrimary>+ Add</ButtonPrimary>
    </Flex>
  );
}

const AddCell = ({ onAdd }: { onAdd: () => void }) => {
  return (
    <Cell align="right">
      <ButtonPrimary onClick={onAdd}>+ Add</ButtonPrimary>
    </Cell>
  );
};
