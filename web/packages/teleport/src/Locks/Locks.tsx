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

import React, { useCallback, useEffect, useState } from 'react';

import Table, { Cell } from 'design/DataTable';
import { ButtonPrimary } from 'design/Button';
import { MenuButton, MenuItem } from 'shared/components/MenuAction';

import api from 'teleport/services/api';
import cfg from 'teleport/config';
import useStickyClusterId from 'teleport/useStickyClusterId';

import {
  FeatureBox,
  FeatureHeader,
  FeatureHeaderTitle,
} from 'teleport/components/Layout';

import { NavLink } from 'teleport/components/Router';

import type { CreateLockData, Lock } from './types';

export function useLocks(clusterId: string) {
  const [locks, setLocks] = useState<Lock[]>([]);

  const fetchLocks = useCallback((clusterId: string) => {
    api.get(cfg.getLocksUrl(clusterId)).then(resp => {
      setLocks(resp);
    });
  }, []);

  const createLock = useCallback(
    async (clusterId: string, createLockData: CreateLockData) => {
      await api.put(cfg.getLocksUrl(clusterId), createLockData);
    },
    []
  );

  useEffect(() => {
    fetchLocks(clusterId);
  }, [clusterId, fetchLocks]);

  return { createLock, fetchLocks, locks };
}

export function Locks() {
  const { clusterId } = useStickyClusterId();
  const { locks, fetchLocks } = useLocks(clusterId);

  function onDelete(lockName: string) {
    api.delete(cfg.getLocksUrlWithUUID(clusterId, lockName)).then(() => {
      // It takes longer for the cache to be updated when removing locks so
      // this waits 1s before fetching the list again.
      setTimeout(() => {
        fetchLocks(clusterId);
      }, 1000);
    });
  }

  return (
    <FeatureBox>
      <FeatureHeader>
        <FeatureHeaderTitle>Session & Identity Locks</FeatureHeaderTitle>
        <ButtonPrimary
          as={NavLink}
          to={cfg.getNewLocksRoute(clusterId)}
          ml="auto"
        >
          + Add New Lock
        </ButtonPrimary>
      </FeatureHeader>
      <Table
        data={locks}
        columns={[
          {
            altKey: 'targets[type]',
            headerText: 'Type',
            isSortable: true,
            render: ({ targets }) => {
              const keys = Object.keys(targets);
              return <Cell>{keys}</Cell>;
            },
          },
          {
            altKey: 'targets[type] value',
            headerText: 'Name',
            isSortable: true,
            render: ({ targets }) => {
              const entries = Object.entries(targets);
              return <Cell>{entries.map(entry => entry[1])}</Cell>;
            },
          },
          {
            key: 'expires',
            headerText: 'Expires',
            isSortable: true,
            render: ({ expires }) => <Cell>{expires}</Cell>,
          },
          {
            key: 'message',
            headerText: 'Reason',
            isSortable: true,
            render: ({ message }) => <Cell>{message}</Cell>,
          },
          {
            altKey: 'options-btn',
            render: ({ name }) => (
              <ManageCell onDelete={onDelete.bind(null, name)} />
            ),
          },
        ]}
        emptyText="No Locks Found"
        isSearchable
        pagination={{ pageSize: 20 }}
      />
    </FeatureBox>
  );
}

const ManageCell = ({ onDelete }: { onDelete: () => void }) => {
  return (
    <Cell align="right">
      <MenuButton>
        <MenuItem onClick={() => onDelete()}>Delete...</MenuItem>
      </MenuButton>
    </Cell>
  );
};
