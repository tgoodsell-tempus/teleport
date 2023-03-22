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

import React, {
  ReactElement,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import styled from 'styled-components';

import { Attempt } from 'shared/hooks/useAsync';

import LinearProgress from 'teleterm/ui/components/LinearProgress';

type ResultListProps<T> = {
  attempt: Attempt<T[]>;
  /**
   * extraItems is an array of extra results that get render irrelevant of the attempt status.
   */
  extraItems?: T[];
  onPick(item: T): void;
  onBack(): void;
  render(item: T): ReactElement;
};

export function ResultList<T>(props: ResultListProps<T>) {
  const { attempt, extraItems = [], onPick, onBack } = props;
  const activeItemRef = useRef<HTMLDivElement>();
  const [activeItemIndex, setActiveItemIndex] = useState(0);

  const items = useMemo(
    () =>
      attempt.status === 'success'
        ? [...extraItems, ...attempt.data]
        : extraItems,
    [attempt.status, attempt.data, extraItems]
  );

  useEffect(() => {
    const handleArrowKey = (e: KeyboardEvent, nudge: number) => {
      const next = getNext(activeItemIndex + nudge, items.length);
      setActiveItemIndex(next);
      // `false` - bottom of the element will be aligned to the bottom of the visible area of the scrollable ancestor
      activeItemRef.current?.scrollIntoView(false);
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'Enter': {
          e.stopPropagation();
          e.preventDefault();

          const item = items[activeItemIndex];
          if (item) {
            onPick(item);
          }
          break;
        }
        case 'Escape': {
          onBack();
          break;
        }
        case 'ArrowUp':
          e.stopPropagation();
          e.preventDefault();

          handleArrowKey(e, -1);
          break;
        case 'ArrowDown':
          e.stopPropagation();
          e.preventDefault();

          handleArrowKey(e, 1);
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [items, attempt.status, onPick, onBack, activeItemIndex]);

  return (
    <>
      {attempt.status === 'processing' && (
        <div
          style={{
            position: 'absolute',
            top: 0,
            height: '1px',
            left: 0,
            right: 0,
          }}
        >
          <LinearProgress transparentBackground={true} />
        </div>
      )}
      {items.map((r, index) => {
        const isActive = index === activeItemIndex;

        return (
          <StyledItem
            ref={isActive ? activeItemRef : null}
            $active={isActive}
            // TODO: Provide a real key instead of using index.
            key={`${index}`}
            onClick={() => props.onPick(r)}
          >
            {props.render(r)}
          </StyledItem>
        );
      })}
    </>
  );
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
    },

    borderBottom: `2px solid ${theme.colors.primary.main}`,
    padding: `${theme.space[1]}px ${theme.space[2]}px`,
    color: theme.colors.primary.contrastText,
    background: $active
      ? theme.colors.primary.lighter
      : theme.colors.primary.light,
  };
});

function getNext(selectedIndex = 0, max = 0) {
  let index = selectedIndex % max;
  if (index < 0) {
    index += max;
  }
  return index;
}
