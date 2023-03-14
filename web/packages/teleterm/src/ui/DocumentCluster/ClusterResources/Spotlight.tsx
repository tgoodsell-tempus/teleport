/**
 * Copyright 2023 Gravitational, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useRef } from 'react';
import { useAsync } from 'shared/hooks/useAsync';
import { ButtonLink, Flex, Text } from 'design';

import { useAppContext } from 'teleterm/ui/appContextProvider';
import { useClusterContext } from 'teleterm/ui/DocumentCluster/clusterContext';
import { Input } from 'teleterm/ui/QuickInput/QuickInput';
import {
  ComponentMap,
  StyledItem,
} from 'teleterm/ui/QuickInput/QuickInputList/QuickInputList';
import { routing } from 'teleterm/ui/uri';

export function Spotlight() {
  const {
    clustersService,
    resourcesService,
    notificationsService,
    workspacesService,
    connectionTracker,
  } = useAppContext();
  clustersService.useState();
  const clusterCtx = useClusterContext();
  const refInput = useRef<HTMLInputElement>();

  const viewAllResults = () => {
    clusterCtx.changeLocation('/resources/servers');
  };

  const [searchAttempt, search] = useAsync(async (inputValue: string) => {
    // TODO: Safely escape this.
    const searchTerms = inputValue
      .split(/\s+/)
      .filter(Boolean)
      .map(term => `"${term}"`)
      .join(',');

    const results = await Promise.all([
      resourcesService
        .fetchServers({
          clusterUri: clusterCtx.clusterUri,
          // TODO: Safely escape this.
          query: `search(${searchTerms})`,
          limit: 10,
        })
        .then(res =>
          res.agentsList.map(server => ({
            kind: 'suggestion.server' as const,
            token: server.hostname,
            data: server,
          }))
        ),
      resourcesService
        .fetchDatabases({
          clusterUri: clusterCtx.clusterUri,
          query: `search(${searchTerms})`,
          limit: 10,
        })
        .then(res =>
          res.agentsList.map(db => ({
            kind: 'suggestion.database' as const,
            token: db.name,
            data: db,
          }))
        ),
      resourcesService
        .fetchKubes({
          clusterUri: clusterCtx.clusterUri,
          query: `search(${searchTerms})`,
          limit: 10,
        })
        .then(res =>
          res.agentsList.map(kube => ({
            kind: 'suggestion.kube' as const,
            token: kube.name,
            data: kube,
          }))
        ),
    ]);

    return results.flat();
  });

  return (
    <Flex flexDirection="column" alignItems="center" gap={3}>
      <div
        css={`
          height: 40px;
          margin: auto;
          width: 300px;
          flex-shrink: 0;
        `}
      >
        <Input
          ref={refInput}
          spellCheck={false}
          placeholder="Spotlight Search"
          onChange={event => {
            search(event.target.value);
          }}
        />
      </div>
      {searchAttempt.status === 'processing' && <Text>Loading</Text>}
      {searchAttempt.status === 'success' && !searchAttempt.data.length && (
        <Text>No results</Text>
      )}
      {searchAttempt.status === 'success' && searchAttempt.data.length > 0 && (
        <div
          css={`
            max-width: 480px;
            margin: 0 auto;
          `}
        >
          {searchAttempt.data.map(item => {
            const Cmpt = ComponentMap[item.kind];
            const onSelect = async () => {
              const resourceUri = item.data.uri;
              const rootClusterUri = routing.ensureRootClusterUri(resourceUri);
              const documentsService =
                workspacesService.getWorkspaceDocumentService(rootClusterUri);

              const connectionToReuse =
                connectionTracker.findConnectionByResourceUri(resourceUri);

              if (connectionToReuse) {
                connectionTracker.activateItem(connectionToReuse.id);
                return;
              }

              await workspacesService.setActiveWorkspace(rootClusterUri);

              switch (item.kind) {
                case 'suggestion.server': {
                  const server = item.data;
                  const doc = documentsService.createTshNodeDocument(
                    server.uri
                  );
                  const rootCluster = clustersService.findClusterByResource(
                    server.uri
                  );
                  // Filer out username for testing purposes.
                  const username = rootCluster?.loggedInUser?.name;
                  const login = rootCluster?.loggedInUser?.sshLoginsList.find(
                    login => login !== username
                  );
                  if (!login) {
                    notificationsService.notifyError(
                      'Could not establish the login for the server'
                    );
                    return;
                  }

                  doc.login = login;
                  doc.title = `${login}@${server.hostname}`;
                  documentsService.add(doc);
                  documentsService.open(doc.uri);
                  break;
                }
                case 'suggestion.database': {
                  const db = item.data;
                  const users = await resourcesService.getDbUsers(db.uri);
                  const user = users[0];

                  if (!user) {
                    notificationsService.notifyError(
                      'Could not establish the user for the database'
                    );
                    return;
                  }

                  const doc = documentsService.createGatewayDocument({
                    targetUri: db.uri,
                    targetName: db.name,
                    // TODO: This has to reuse logic from useDatabases.
                    targetUser: user,
                  });
                  documentsService.add(doc);
                  documentsService.open(doc.uri);
                  break;
                }
                case 'suggestion.kube': {
                  // TODO: Use correct cluster to connect kube.
                  clusterCtx.connectKube(item.data.uri);
                  return;
                }
              }
            };

            return (
              <StyledItem key={item.data.uri} onClick={onSelect}>
                <Cmpt item={item} />
              </StyledItem>
            );
          })}
          <ButtonLink type="button" onClick={viewAllResults}>
            View all results
          </ButtonLink>
        </div>
      )}
    </Flex>
  );
}
