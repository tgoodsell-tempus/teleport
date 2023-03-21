export const HOOK_LIST = [
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

export const HOOK_CREATED = {
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
