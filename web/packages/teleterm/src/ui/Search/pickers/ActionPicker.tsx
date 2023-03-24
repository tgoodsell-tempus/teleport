import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import styled from 'styled-components';

import { Box, Flex, Label as DesignLabel, Text } from 'design';
import * as icons from 'design/Icon';

import { makeEmptyAttempt, useAsync, mapAttempt } from 'shared/hooks/useAsync';
import { Highlight } from 'shared/components/Highlight';

import Logger from 'teleterm/logger';
import { useAppContext } from 'teleterm/ui/appContextProvider';
import {
  ResourceMatch,
  SearchResult,
  SearchResultDatabase,
  SearchResultKube,
  SearchResultServer,
} from 'teleterm/ui/Search/searchResult';
import * as tsh from 'teleterm/services/tshd/types';
import { sortResults, useSearch } from 'teleterm/ui/Search/useSearch';
import * as uri from 'teleterm/ui/uri';

import { mapToActions, SearchAction } from '../actions';
import { useSearchContext } from '../SearchContext';

import { getParameterPicker } from './pickers';
import { ResultList, EmptyListCopy } from './ResultList';

export function ActionPicker() {
  const searchLogger = useRef(new Logger('search'));
  const ctx = useAppContext();
  const { clustersService } = ctx;

  const [searchAttempt, fetch, setAttempt] = useAsync(useSearch());
  const { inputValue, changeActivePicker, close } = useSearchContext();
  const debouncedInputValue = useDebounce(inputValue, 200);

  const attempt = useMemo(
    () =>
      mapAttempt(searchAttempt, ({ results, search }) => {
        const sortedResults = sortResults(results, search);
        searchLogger.current.info('results for', search, sortedResults);

        return mapToActions(ctx, sortedResults);
      }),
    [ctx, searchAttempt]
  );

  const getClusterName = useCallback(
    (resourceUri: uri.ResourceUri) => {
      const clusterUri = uri.routing.ensureClusterUri(resourceUri);
      const cluster = clustersService.findCluster(clusterUri);

      return cluster ? cluster.name : uri.routing.parseClusterName(resourceUri);
    },
    [clustersService]
  );

  useEffect(() => {
    if (debouncedInputValue) {
      fetch(debouncedInputValue);
    } else {
      setAttempt(makeEmptyAttempt());
    }
  }, [debouncedInputValue]);

  const onPick = useCallback(
    (action: SearchAction) => {
      if (action.type === 'simple-action') {
        action.perform();
        close();
      }
      if (action.type === 'parametrized-action') {
        changeActivePicker(getParameterPicker(action));
      }
    },
    [changeActivePicker, close]
  );

  if (!inputValue) {
    return (
      <EmptyListCopy>
        <Text>
          <ul>
            <li>Separate the search terms with space.</li>
            <li>
              Resources that match the query the most will appear at the top.
            </li>
            <li>
              Selecting a search result will connect to the resource in a new
              tab.
            </li>
          </ul>
        </Text>
      </EmptyListCopy>
    );
  }

  return (
    <ResultList<SearchAction>
      attempt={attempt}
      onPick={onPick}
      onBack={close}
      render={item => {
        const Component = ComponentMap[item.searchResult.kind];
        return (
          <Component
            searchResult={item.searchResult}
            getClusterName={getClusterName}
          />
        );
      }}
    />
  );
}

function useDebounce<T>(value: T, delay: number): T {
  // State and setters for debounced value
  const [debouncedValue, setDebouncedValue] = useState(value);
  useEffect(
    () => {
      // Update debounced value after delay
      const handler = setTimeout(() => setDebouncedValue(value), delay);
      // Cancel the timeout if value changes (also on delay change or unmount)
      // This is how we prevent debounced value from updating if value is changed ...
      // .. within the delay period. Timeout gets cleared and restarted.
      return () => clearTimeout(handler);
    },
    [value, delay] // Only re-call effect if value or delay changes
  );
  return debouncedValue;
}

const ComponentMap: Record<
  SearchResult['kind'],
  React.FC<SearchResultItem<SearchResult>>
> = {
  server: ServerItem,
  kube: KubeItem,
  database: DatabaseItem,
};

type SearchResultItem<T> = {
  searchResult: T;
  getClusterName: (uri: uri.ResourceUri) => string;
};

function ServerItem(props: SearchResultItem<SearchResultServer>) {
  const { searchResult } = props;
  const server = searchResult.resource;

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
              <HighlightField field="hostname" searchResult={searchResult} />
            </strong>
          </Text>
        </Flex>
        <Box>
          <Text typography="body2" fontSize={0}>
            {props.getClusterName(server.uri)}
          </Text>
        </Box>
      </Flex>

      <Labels searchResult={searchResult}>
        <DesignLabel key={'addr'} kind="secondary">
          {server.tunnel ? (
            <span title="This node is connected to the cluster through a reverse tunnel">
              ↵ tunnel
            </span>
          ) : (
            <HighlightField field="addr" searchResult={searchResult} />
          )}
        </DesignLabel>
      </Labels>
    </Flex>
  );
}

function DatabaseItem(props: SearchResultItem<SearchResultDatabase>) {
  const { searchResult } = props;
  const db = searchResult.resource;

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
              <HighlightField field="name" searchResult={searchResult} />
            </strong>
          </Text>
        </Flex>
        <Box>
          <Text typography="body2" fontSize={0}>
            {props.getClusterName(db.uri)}
          </Text>
        </Box>
      </Flex>

      <Labels searchResult={searchResult}>
        <DesignLabel key={'type-protocol'} kind="secondary">
          <HighlightField field="type" searchResult={searchResult} />
          /
          <HighlightField field="protocol" searchResult={searchResult} />
        </DesignLabel>
        {db.desc && (
          <DesignLabel key={'desc'} kind="secondary">
            <HighlightField field="desc" searchResult={searchResult} />
          </DesignLabel>
        )}
      </Labels>
    </Flex>
  );
}

function KubeItem(props: SearchResultItem<SearchResultKube>) {
  const { searchResult } = props;

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
              <HighlightField field="name" searchResult={searchResult} />
            </strong>
          </Text>
        </Flex>
        <Box>
          <Text typography="body2" fontSize={0}>
            {props.getClusterName(searchResult.resource.uri)}
          </Text>
        </Box>
      </Flex>

      <Labels searchResult={searchResult} />
    </Flex>
  );
}

function Labels(
  props: React.PropsWithChildren<{ searchResult: SearchResult }>
) {
  const { searchResult } = props;

  return (
    <Flex gap={1} flexWrap="wrap">
      {props.children}
      {searchResult.resource.labelsList.map(label => (
        <Label
          key={label.name + label.value}
          searchResult={searchResult}
          label={label}
        />
      ))}
    </Flex>
  );
}

function Label(props: { searchResult: SearchResult; label: tsh.Label }) {
  const { searchResult: item, label } = props;
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
