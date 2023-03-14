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

type Lock = {
  name: string;
  message: string;
  expires: string;
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

export function useLocks(clusterId: string) {
  const [locks, setLocks] = useState<Lock[]>([]);

  useEffect(() => {
    api.get(cfg.getLocksUrl(clusterId)).then(resp => {
      setLocks(resp);
    });
  }, [clusterId]);

  return locks;
}

export function Locks() {
  const { clusterId } = useStickyClusterId();
  const lockData = useLocks(clusterId);

  function onDelete(lockName: string) {
    console.log('deleting', lockName);
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
        data={lockData}
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
            render: ({ name }) => <ManageCell onDelete={onDelete.bind(name)} />,
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
