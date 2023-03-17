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

import React, { useRef } from 'react';
import SlidePanel from 'design/SlidePanel';

import { ArrowBack, Trash } from 'design/Icon';
import { Box, ButtonPrimary, Flex, Input, Text } from 'design';
import { Cell } from 'design/DataTable';

import useStickyClusterId from 'teleport/useStickyClusterId';
import history from 'teleport/services/history';
import cfg from 'teleport/config';

import { useLocks } from './Locks';
import { StyledTable } from './shared';

import type { Positions } from 'design/SlidePanel/SlidePanel';
import type { CreateLockData, SelectedLockTarget } from './types';

type Props = {
  panelPosition: Positions;
  setPanelPosition: (Positions) => void;
  selectedLockTargets: SelectedLockTarget[];
  setSelectedLockTargets: (lockTargets: SelectedLockTarget[]) => void;
};

export function CreateLock({
  panelPosition,
  setPanelPosition,
  selectedLockTargets,
  setSelectedLockTargets,
}: Props) {
  const { clusterId } = useStickyClusterId();
  const { createLock } = useLocks(clusterId);

  const messageRef = useRef<HTMLInputElement>(null);
  const ttlRef = useRef<HTMLInputElement>(null);

  function handleCreateLock() {
    selectedLockTargets.forEach(async lockTarget => {
      const lockData: CreateLockData = {
        targets: { [lockTarget.type]: lockTarget.name },
      };
      const message = messageRef?.current?.value;
      const ttl = ttlRef?.current?.value;
      if (message) lockData.message = message;
      if (ttl) lockData.ttl = ttl;
      await createLock(clusterId, lockData);
    });
    setTimeout(() => {
      // It takes longer for the cache to be updated when adding locks so
      // this waits 1s before redirecting to fetch the list again.
      history.push(cfg.getLocksRoute(clusterId));
    }, 1000);
  }

  function onRemove(name) {
    const index = selectedLockTargets.findIndex(target => target.name === name);
    selectedLockTargets.splice(index, 1);
    setSelectedLockTargets([...selectedLockTargets]);
  }

  return (
    <SlidePanel
      position={panelPosition}
      closePanel={() => setPanelPosition('closed')}
    >
      <div>
        <Flex alignItems="center">
          <ArrowBack
            fontSize={25}
            mr={3}
            onClick={() => setPanelPosition('closed')}
            style={{ cursor: 'pointer' }}
          />
          <Box>
            <Text typography="h4" color="light" bold>
              Create New Lock
            </Text>
          </Box>
        </Flex>

        <StyledTable
          data={selectedLockTargets}
          css={`
            margin-top: 3rem;
          `}
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
            {
              altKey: 'remove-btn',
              render: ({ name }) => (
                <Cell align="right">
                  <Trash
                    fontSize={13}
                    borderRadius={2}
                    p={2}
                    onClick={onRemove.bind(null, name)}
                    css={`
                      cursor: pointer;
                      background-color: #2e3860;
                      border-radius: 2px;
                      :hover {
                        background-color: #414b70;
                      }
                    `}
                  />
                </Cell>
              ),
            },
          ]}
          emptyText="No Targets Found"
        />
        <Box mt={3}>
          <Text mr={2}>Message: </Text>
          <Input placeholder={`Going down for maintenance`} ref={messageRef} />
        </Box>
        <Box mt={3}>
          <Text mr={2}>TTL: </Text>
          <Input placeholder={`5h`} ref={ttlRef} />
        </Box>
      </div>
      <Flex mt={5} justifyContent="flex-end">
        <ButtonPrimary
          width="165px"
          onClick={handleCreateLock}
          disabled={!selectedLockTargets.length}
        >
          Create locks
        </ButtonPrimary>
      </Flex>
    </SlidePanel>
  );
}
