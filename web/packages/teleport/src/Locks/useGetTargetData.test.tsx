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

import { renderHook } from '@testing-library/react-hooks';

import { useGetTargetData } from './useGetTargetData';

const mockedUtils = {
  mfaService: {
    fetchDevices: () =>
      new Promise(resolve =>
        resolve([
          {
            id: '4bac1adb-fdaa-4c31-a989-317892a9d1bd',
            name: 'yubikey',
            description: 'Hardware Key',
            registeredDate: '2023-03-14T19:22:59.437Z',
            lastUsedDate: '2023-03-21T19:03:54.874Z',
          },
        ])
      ),
  },
  desktopService: {
    fetchDesktops: () =>
      new Promise(resolve =>
        resolve({
          agents: [
            {
              os: 'windows',
              name: 'watermelon',
              addr: 'localhost.watermelon',
              labels: [
                {
                  name: 'env',
                  value: 'test',
                },
                {
                  name: 'os',
                  value: 'os',
                },
                {
                  name: 'unique-id',
                  value: '47c38f49-b690-43fd-ac28-946e7a0a6188',
                },
                {
                  name: 'windows-desktops',
                  value: 'watermelon',
                },
              ],
              host_id: '47c38f49-b690-43fd-ac28-946e7a0a6188',
              logins: [],
            },
            {
              os: 'windows',
              name: 'banana',
              addr: 'localhost.banana',
              labels: [
                {
                  name: 'env',
                  value: 'test',
                },
                {
                  name: 'os',
                  value: 'linux',
                },
                {
                  name: 'unique-id',
                  value: '4c3bd959-8444-492a-a383-a29378da93c9',
                },
                {
                  name: 'windows-desktops',
                  value: 'banana',
                },
              ],
              host_id: '4c3bd959-8444-492a-a383-a29378da93c9',
              logins: [],
            },
          ],
          startKey: '',
          totalCount: 0,
        })
      ),
  },
  nodeService: {
    fetchNodes: () =>
      new Promise(resolve =>
        resolve({
          agents: [
            {
              id: 'e14baac6-15c1-42c2-a7d9-99410d21cf4c',
              clusterId: 'local-test2',
              hostname: 'node1.go.citadel',
              labels: ['special:apple', 'user:orange'],
              addr: '127.0.0.1:4022',
              tunnel: false,
              sshLogins: [],
            },
          ],
          startKey: '',
          totalCount: 0,
        })
      ),
  },
  resourceService: {
    fetchRoles: () =>
      new Promise(resolve =>
        resolve([
          {
            id: 'role:admin',
            kind: 'role',
            name: 'admin',
            content: '',
          },
          {
            id: 'role:contractor',
            kind: 'role',
            name: 'contractor',
            content: '',
          },
          {
            id: 'role:locksmith',
            kind: 'role',
            name: 'locksmith',
            content: '',
          },
        ])
      ),
  },
  userService: {
    fetchUsers: () =>
      new Promise(resolve =>
        resolve([
          {
            name: 'admin-local',
            roles: ['access', 'admin', 'auditor', 'editor'],
            authType: 'local',
          },
          {
            name: 'admin',
            roles: ['access', 'admin', 'auditor', 'editor', 'locksmith'],
            authType: 'local',
          },
          {
            name: 'worker',
            roles: ['access', 'contractor'],
            authType: 'local',
          },
        ])
      ),
  },
};

const additionalTargets = {
  access_request: {
    fetch: () =>
      new Promise(resolve =>
        resolve([
          {
            name: 'apple',
            description: 'tree',
            date: '1/2/1234',
          },
        ])
      ),
    handler: (setter, requests) => {
      const filteredData = requests.map(r => ({
        name: r.name,
        description: r.description,
        theDate: r.date,
      }));
      setter(filteredData);
    },
    options: {},
  },
};

jest.mock('teleport/useTeleport', () => ({
  __esModule: true,
  default: () => mockedUtils,
}));

describe('hook: useLocks', () => {
  describe('can fetch and filter', () => {
    it('mfa data', async () => {
      async () => {
        const { result, waitForNextUpdate } = renderHook(() =>
          useGetTargetData('windows_desktop', 'cluster-id')
        );
        await waitForNextUpdate();
        expect(result.current).toStrictEqual([
          {
            name: 'yubikey',
            id: '4bac1adb-fdaa-4c31-a989-317892a9d1bd',
            description: 'Hardware Key',
            lastUsed: 'Tue, 21 Mar 2023 19:03:54 GMT',
          },
        ]);
      };
    });

    it('desktops data', async () => {
      const { result, waitForNextUpdate } = renderHook(() =>
        useGetTargetData('windows_desktop', 'cluster-id')
      );
      await waitForNextUpdate();
      expect(result.current).toStrictEqual([
        {
          name: 'watermelon',
          addr: 'localhost.watermelon',
          labels:
            'env:test, os:os, unique-id:47c38f49-b690-43fd-ac28-946e7a0a6188, windows-desktops:watermelon',
        },
        {
          name: 'banana',
          addr: 'localhost.banana',
          labels:
            'env:test, os:linux, unique-id:4c3bd959-8444-492a-a383-a29378da93c9, windows-desktops:banana',
        },
      ]);
    });

    it('nodes data', async () => {
      const { result, waitForNextUpdate } = renderHook(() =>
        useGetTargetData('node', 'cluster-id')
      );
      await waitForNextUpdate();
      expect(result.current).toStrictEqual([
        {
          name: 'node1.go.citadel',
          addr: '127.0.0.1:4022',
          labels: 'special:apple, user:orange',
        },
      ]);
    });

    it('roles data', async () => {
      const { result, waitForNextUpdate } = renderHook(() =>
        useGetTargetData('role', 'cluster-id')
      );
      await waitForNextUpdate();
      expect(result.current).toStrictEqual([
        { name: 'admin' },
        { name: 'contractor' },
        { name: 'locksmith' },
      ]);
    });

    it('user data', async () => {
      const { result, waitForNextUpdate } = renderHook(() =>
        useGetTargetData('user', 'cluster-id')
      );
      await waitForNextUpdate();
      expect(result.current).toStrictEqual([
        { name: 'admin-local', roles: 'access, admin, auditor, editor' },
        { name: 'admin', roles: 'access, admin, auditor, editor, locksmith' },
        { name: 'worker', roles: 'access, contractor' },
      ]);
    });

    it('additionally supplied targets', async () => {
      const { result, waitForNextUpdate } = renderHook(() =>
        useGetTargetData('access_request', 'cluster-id', additionalTargets)
      );
      await waitForNextUpdate();
      expect(result.current).toStrictEqual([
        { name: 'apple', description: 'tree', theDate: '1/2/1234' },
      ]);
    });
  });
});
