/**
 * Copyright 2023 Gravitational, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useEffect, useRef } from 'react';
import styled from 'styled-components';
import { Box, Text, Flex, Label as DesignLabel } from 'design';
import * as icons from 'design/Icon';
import { Highlight } from 'shared/components/Highlight';

import * as types from 'teleterm/ui/services/searchBar/types';
import {
  ActionDbConnect,
  ActionKubeConnect,
  ActionSshConnect,
  SearchBarAction,
} from 'teleterm/ui/services/searchBar/types';
import { SearchResult, ResourceMatch } from 'teleterm/ui/Search/searchResult';
import * as uri from 'teleterm/ui/uri';

import type * as tsh from 'teleterm/services/tshd/types';
import type { Attempt } from 'shared/hooks/useAsync';

type Props = {
  attempt: Attempt<types.SearchBarAction[]>;
  activeItem: number;
  isPristine: boolean;
  onPick(item: types.SearchBarAction): void;
};

export const SearchBarList = React.forwardRef<HTMLElement, Props>(
  (props, ref) => {
    const { attempt, activeItem, isPristine } = props;
    const activeItemRef = useRef<HTMLDivElement>();

    useEffect(() => {
      // `false` - bottom of the element will be aligned to the bottom of the visible area of the scrollable ancestor
      activeItemRef.current?.scrollIntoView(false);
    }, [activeItem]);

    function handleClick(e: React.SyntheticEvent) {
      if (attempt.status !== 'success' || attempt.data.length === 0) {
        return;
      }

      const el = e.target;
      if (el instanceof Element) {
        const itemEl = el.closest('[data-attr]');
        const index = parseInt(itemEl.getAttribute('data-attr'));
        props.onPick(attempt.data[index]);
      }
    }

    return (
      <StyledGlobalSearchResults
        ref={ref}
        tabIndex={-1}
        data-attr="quickpicker.list"
        onClick={handleClick}
      >
        {isPristine ? (
          <EmptyListText>
            <Text>
              <ul>
                <li>Separate the search terms with space.</li>
                <li>
                  Resources that match the query the most will appear at the
                  top.
                </li>
                <li>
                  Selecting a search result will connect to the resource in a
                  new tab.
                </li>
              </ul>
            </Text>
          </EmptyListText>
        ) : (
          attempt.status === 'success' &&
          (attempt.data.length === 0 ? (
            <EmptyListText>
              <Text>No matching items found.</Text>
            </EmptyListText>
          ) : (
            attempt.data.map((r, index) => {
              const Cmpt = ComponentMap[r.kind] || UnknownItem;
              const isActive = index === activeItem;

              return (
                <StyledItem
                  data-attr={index}
                  ref={isActive ? activeItemRef : null}
                  $active={isActive}
                  key={`${index}`}
                >
                  <Cmpt item={r} />
                </StyledItem>
              );
            })
          ))
        )}
      </StyledGlobalSearchResults>
    );
  }
);

function UnknownItem(props: { item: types.SearchBarAction }) {
  const { kind } = props.item;
  return <div>unknown kind: {kind} </div>;
}

function SshLoginItem(props: { item: types.ActionSshLogin }) {
  return <div>{props.item.searchResult.login}</div>;
}

function DbUsernameItem(props: { item: types.ActionDbUsername }) {
  return <div>{props.item.searchResult.username}</div>;
}

const StyledItem = styled.div(({ theme, $active }) => {
  return {
    '&:hover, &:focus': {
      cursor: 'pointer',
      background: theme.colors.primary.lighter,
    },
    '& mark': {
      color: 'inherit',
      backgroundColor: theme.colors.secondary.light,
      // backgroundColor: 'inherit',
      // filter: 'invert(100%)',
      // 'mix-blend-mode': 'difference',
    },

    borderBottom: `2px solid ${theme.colors.primary.main}`,
    padding: `${theme.space[1]}px ${theme.space[2]}px`,
    color: theme.colors.primary.contrastText,
    background: $active
      ? theme.colors.primary.lighter
      : theme.colors.primary.light,
  };
});

const StyledGlobalSearchResults = styled.div(({ theme }) => {
  return {
    boxShadow: '8px 8px 18px rgb(0 0 0)',
    color: theme.colors.primary.contrastText,
    background: theme.colors.primary.light,
    boxSizing: 'border-box',
    width: '600px',
    marginTop: '32px',
    display: 'block',
    position: 'absolute',
    border: '1px solid ' + theme.colors.action.hover,
    fontSize: '12px',
    listStyle: 'none outside none',
    textShadow: 'none',
    zIndex: '1000',
    maxHeight: '350px',
    overflow: 'auto',
    minHeight: '50px',
  };
});

const ComponentMap: Record<
  SearchBarAction['kind'],
  React.FC<{ item: SearchBarAction }>
> = {
  ['action.ssh-connect']: ServerItem,
  ['action.kube-connect']: KubeItem,
  ['action.db-connect']: DatabaseItem,
  ['action.ssh-login']: SshLoginItem,
  ['action.db-username']: DbUsernameItem,
};

// TODO(ravicious): Get cluster name from ClustersService.
// TODO(ravicious): Show cluster name only if the user is logged in to more than 1 cluster.
function clusterName(resourceUri: uri.ResourceUri): string {
  return uri.routing.parseClusterName(resourceUri);
}

function ServerItem(props: { item: ActionSshConnect }) {
  const server = props.item.searchResult.resource;
  return (
    <Flex flexDirection="column" minWidth="300px" gap={1}>
      <Flex justifyContent="space-between" alignItems="center">
        <Flex alignItems="center" gap={1} flex="1 0">
          <SquareIconBackground color="#c05b9e">
            <icons.Server />
          </SquareIconBackground>
          <Text typography="body1">
            Connect over SSH to{' '}
            <strong>
              <HighlightField
                field="hostname"
                searchResult={props.item.searchResult}
              />
            </strong>
          </Text>
        </Flex>
        <Box>
          <Text typography="body2" fontSize={0}>
            {clusterName(props.item.searchResult.resource.uri)}
          </Text>
        </Box>
      </Flex>

      <Labels item={props.item.searchResult}>
        <DesignLabel key={'addr'} kind="secondary">
          {server.tunnel ? (
            <span title="This node is connected to the cluster through a reverse tunnel">
              ↵ tunnel
            </span>
          ) : (
            <HighlightField
              field="addr"
              searchResult={props.item.searchResult}
            />
          )}
        </DesignLabel>
      </Labels>
    </Flex>
  );
}

function DatabaseItem(props: { item: ActionDbConnect }) {
  const db = props.item.searchResult.resource;

  return (
    <Flex flexDirection="column" minWidth="300px" gap={1}>
      <Flex justifyContent="space-between" alignItems="center">
        <Flex alignItems="center" gap={1} flex="1 0">
          <SquareIconBackground
            color="#4ab9c9"
            // The database icon is different than ssh and kube icons for some reason.
            css={`
              padding-left: 5px;
              padding-top: 5px;
            `}
          >
            <icons.Database />
          </SquareIconBackground>
          <Text typography="body1">
            Set up a db connection for{' '}
            <strong>
              <HighlightField
                field="name"
                searchResult={props.item.searchResult}
              />
            </strong>
          </Text>
        </Flex>
        <Box>
          <Text typography="body2" fontSize={0}>
            {clusterName(db.uri)}
          </Text>
        </Box>
      </Flex>

      <Labels item={props.item.searchResult}>
        <DesignLabel key={'type-protocol'} kind="secondary">
          <HighlightField field="type" searchResult={props.item.searchResult} />
          /
          <HighlightField
            field="protocol"
            searchResult={props.item.searchResult}
          />
        </DesignLabel>
        {db.desc && (
          <DesignLabel key={'desc'} kind="secondary">
            <HighlightField
              field="desc"
              searchResult={props.item.searchResult}
            />
          </DesignLabel>
        )}
      </Labels>
    </Flex>
  );
}

function KubeItem(props: { item: ActionKubeConnect }) {
  return (
    <Flex flexDirection="column" minWidth="300px" gap={1}>
      <Flex justifyContent="space-between" alignItems="center">
        <Flex alignItems="center" gap={1} flex="1 0">
          <SquareIconBackground color="#326ce5">
            <icons.Kubernetes />
          </SquareIconBackground>
          <Text typography="body1">
            Log in to Kubernetes cluster{' '}
            <strong>
              <HighlightField
                field="name"
                searchResult={props.item.searchResult}
              />
            </strong>
          </Text>
        </Flex>
        <Box>
          <Text typography="body2" fontSize={0}>
            {clusterName(props.item.searchResult.resource.uri)}
          </Text>
        </Box>
      </Flex>

      <Labels item={props.item.searchResult} />
    </Flex>
  );
}

function Labels(props: React.PropsWithChildren<{ item: SearchResult }>) {
  return (
    <Flex gap={1} flexWrap="wrap">
      {props.children}
      {props.item.resource.labelsList.map(label => (
        <Label key={label.name + label.value} item={props.item} label={label} />
      ))}
    </Flex>
  );
}

function Label(props: { item: SearchResult; label: tsh.Label }) {
  const { item, label } = props;
  const labelMatches = item.labelMatches.filter(
    match => match.labelName == label.name
  );
  const nameMatches = labelMatches
    .filter(match => match.kind === 'label-name')
    .map(match => match.searchTerm);
  const valueMatches = labelMatches
    .filter(match => match.kind === 'label-value')
    .map(match => match.searchTerm);

  return (
    <DesignLabel key={label.name} kind="secondary">
      <Highlight text={label.name} keywords={nameMatches} />:{' '}
      <Highlight text={label.value} keywords={valueMatches} />
    </DesignLabel>
  );
}

function HighlightField(props: {
  searchResult: SearchResult;
  field: ResourceMatch<SearchResult['kind']>['field'];
}) {
  // `as` used as a workaround for a TypeScript issue.
  // https://github.com/microsoft/TypeScript/issues/33591
  const keywords = (
    props.searchResult.resourceMatches as ResourceMatch<SearchResult['kind']>[]
  )
    .filter(match => match.field === props.field)
    .map(match => match.searchTerm);

  return (
    <Highlight
      text={props.searchResult.resource[props.field]}
      keywords={keywords}
    />
  );
}

const SquareIconBackground = styled(Box)`
  background: ${props => props.color};
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 24px;
  width: 24px;
  border-radius: 2px;
  padding: 4px;
  font-size: 18px;
`;

const EmptyListText = styled(Box)`
  width: 100%;
  height: 100%;
  padding: ${props => props.theme.space[2]}px;
  line-height: 1.5em;

  ul {
    margin: 0;
    padding-inline-start: 2em;
  }
`;
