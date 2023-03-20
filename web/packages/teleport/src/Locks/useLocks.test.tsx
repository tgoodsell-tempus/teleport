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

import { useLocks } from './useLocks';

jest.mock('teleport/services/api', () => ({
  get: () => new Promise(resolve => resolve(HOOK_LIST)),
  put: () => new Promise(resolve => resolve(HOOK_CREATED)),
}));

describe('hook: useLocks', () => {
  it('fetches and returns the locks', async () => {
    const { result, waitForNextUpdate } = renderHook(() =>
      useLocks('cluster-id')
    );
    result.current.fetchLocks('cluster-id');
    expect(result.current.locks).toHaveLength(0);
    await waitForNextUpdate();
    expect(result.current.locks).toHaveLength(4);
  });

  it('creates locks', async () => {
    const { result, waitForNextUpdate } = renderHook(() =>
      useLocks('cluster-id')
    );
    // When the hook is initialized it fetches all hooks so wait for this to
    // happen before continuing on.
    await waitForNextUpdate();
    const resp = await result.current.createLock('cluster-id', {
      targets: { user: 'banned' },
      message: "you've been bad",
      ttl: '5h',
    });
    expect(resp).toBe(HOOK_CREATED);
  });
});

const HOOK_LIST = [
  {
    name: '1ecfe67f-a59b-4309-b6fc-a9981891e82a',
    message: "you've been bad",
    expires: '2023-03-18T02:14:01.659948Z',
    createdAt: '',
    createdBy: '',
    targets: { user: 'worker' },
  },
  {
    name: '3ec76143-1ebb-4328-acbb-83799919e2a8',
    message: 'Forever gone',
    expires: '',
    createdAt: '2023-03-20T16:57:17.117411Z',
    createdBy: 'tele-admin-local',
    targets: { user: 'worker' },
  },
  {
    name: '5df33ee0-6368-4f9d-b8d1-9a4121830018',
    message: '',
    expires: '2023-03-20T21:10:17.529834Z',
    createdAt: '2023-03-20T19:10:17.533992Z',
    createdBy: 'tele-admin-local',
    targets: { user: 'worker' },
  },
  {
    name: '60626e99-e91b-41b2-89fe-bf5d16b0c622',
    message: 'No contractors allowed right now',
    expires: '2023-03-20T19:36:15.028132Z',
    createdAt: '2023-03-20T14:36:15.046728Z',
    createdBy: 'tele-admin',
    targets: { role: 'contractor' },
  },
];

const HOOK_CREATED = {
  kind: 'lock',
  version: 'v2',
  metadata: { name: '1b807c9f-2144-4f7f-8d3e-9c1e14cb5b98' },
  spec: {
    target: { user: 'banned' },
    message: "you've been bad",
    expires: '2023-03-20T21:51:18.466627Z',
    created_at: '0001-01-01T00:00:00Z',
  },
};
