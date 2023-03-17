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
import styled from 'styled-components';
import * as icons from 'design/Icon';
import { useAsync } from 'shared/hooks/useAsync';
import { ButtonLink, Flex, Text, Box, Label as DesignLabel } from 'design';
import { Highlight } from 'shared/components/Highlight';

import { useAppContext } from 'teleterm/ui/appContextProvider';
import { useClusterContext } from 'teleterm/ui/DocumentCluster/clusterContext';
import { Input } from 'teleterm/ui/QuickInput/QuickInput';
import { StyledItem } from 'teleterm/ui/QuickInput/QuickInputList/QuickInputList';
import { useSearch, sortResults } from 'teleterm/ui/Search/useSearch';
import { routing } from 'teleterm/ui/uri';

import type * as types from 'teleterm/ui/services/resources';
import type * as tsh from 'teleterm/services/tshd/types';

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

  const [searchAttempt, search] = useAsync(useSearch());
  console.log('searchAttempt data', searchAttempt.data);

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
      {searchAttempt.status === 'success' &&
        !searchAttempt.data.results.length && <Text>No results</Text>}
      {searchAttempt.status === 'success' &&
        searchAttempt.data.results.length > 0 && (
          <div
            css={`
              max-width: 480px;
              margin: 0 auto;
            `}
          >
            {sortResults(
              searchAttempt.data.results,
              searchAttempt.data.search
            ).map(searchResult => {
              const Cmpt = ComponentMap[searchResult.kind];
              const onSelect = async () => {
                const resourceUri = searchResult.resource.uri;
                const rootClusterUri =
                  routing.ensureRootClusterUri(resourceUri);
                const documentsService =
                  workspacesService.getWorkspaceDocumentService(rootClusterUri);

                const connectionToReuse =
                  connectionTracker.findConnectionByResourceUri(resourceUri);

                if (connectionToReuse) {
                  connectionTracker.activateItem(connectionToReuse.id);
                  return;
                }

                await workspacesService.setActiveWorkspace(rootClusterUri);

                switch (searchResult.kind) {
                  case 'server': {
                    const server = searchResult.resource;
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
                  case 'database': {
                    const db = searchResult.resource;
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
                  case 'kube': {
                    // TODO: Use correct cluster to connect kube.
                    clusterCtx.connectKube(searchResult.resource.uri);
                    return;
                  }
                }
              };

              return (
                <StyledItem key={searchResult.resource.uri} onClick={onSelect}>
                  <Cmpt searchResult={searchResult} />
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

const ComponentMap: Record<
  types.SearchResult['kind'],
  React.FC<{ searchResult: types.SearchResult }>
> = {
  ['server']: ServerItem,
  ['database']: DatabaseItem,
  ['kube']: KubeItem,
};

function ServerItem(props: { searchResult: types.SearchResultServer }) {
  const { hostname } = props.searchResult.resource;

  return (
    <Flex alignItems="flex-start" p={1} minWidth="300px">
      <SquareIconBackground color="#4DB2F0">
        <icons.Server fontSize="20px" />
      </SquareIconBackground>
      <Flex flexDirection="column" ml={1} flex={1}>
        <Flex justifyContent="space-between" alignItems="center">
          <Box mr={2}>{hostname}</Box>
          <Box>
            <Text typography="body2" fontSize={0}>
              {props.searchResult.score}
            </Text>
          </Box>
        </Flex>
        <Labels searchResult={props.searchResult} />
      </Flex>
    </Flex>
  );
}

function DatabaseItem(props: { searchResult: types.SearchResultDatabase }) {
  const db = props.searchResult.resource;

  return (
    <Flex alignItems="flex-start" p={1} minWidth="300px">
      <SquareIconBackground color="#4DB2F0">
        <icons.Database fontSize="20px" />
      </SquareIconBackground>
      <Flex flexDirection="column" ml={1} flex={1}>
        <Flex justifyContent="space-between" alignItems="center">
          <Box mr={2}>{db.name}</Box>
          <Box>
            <Text typography="body2" fontSize={0}>
              {db.type}/{db.protocol} {props.searchResult.score}
            </Text>
          </Box>
        </Flex>
        <Labels searchResult={props.searchResult} />
      </Flex>
    </Flex>
  );
}

function KubeItem(props: { searchResult: types.SearchResultKube }) {
  const { name } = props.searchResult.resource;

  return (
    <Flex alignItems="flex-start" p={1} minWidth="300px">
      <SquareIconBackground color="#4DB2F0">
        <icons.Kubernetes fontSize="20px" />
      </SquareIconBackground>
      <Flex flexDirection="column" ml={1} flex={1}>
        <Box mr={2}>{name}</Box>
        <Labels searchResult={props.searchResult} />
      </Flex>
    </Flex>
  );
}

function Labels(props: { searchResult: types.SearchResult }) {
  return (
    <Flex gap={1} flexWrap="wrap">
      {props.searchResult.resource.labelsList.map(label => (
        <Label
          key={label.name + label.value}
          searchResult={props.searchResult}
          label={label}
        />
      ))}
    </Flex>
  );
}

function Label(props: { searchResult: types.SearchResult; label: tsh.Label }) {
  const { searchResult, label } = props;
  const labelMatches = searchResult.labelMatches.filter(
    match => match.matchedValue.labelName == label.name
  );
  const nameMatches = labelMatches
    .filter(match => match.matchedValue.kind === 'label-name')
    .map(match => match.searchTerm);
  const valueMatches = labelMatches
    .filter(match => match.matchedValue.kind === 'label-value')
    .map(match => match.searchTerm);

  return (
    <DesignLabel key={label.name} kind="secondary">
      <Highlight text={label.name} keywords={nameMatches} />:{' '}
      <Highlight text={label.value} keywords={valueMatches} />
    </DesignLabel>
  );
}

const SquareIconBackground = styled(Box)`
  background: ${props => props.color};
  display: flex;
  align-items: center;
  justify-content: center;
  height: 26px;
  width: 26px;
  margin-right: 8px;
  border-radius: 2px;
  padding: 4px;
`;
