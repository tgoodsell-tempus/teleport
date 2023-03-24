/**
 * Copyright 2022 Gravitational, Inc.
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
  useContext,
  useState,
  FC,
  useCallback,
  createContext,
  useRef,
  MutableRefObject,
} from 'react';

import { actionPicker, SearchPicker } from './pickers/pickers';

const SearchContext =
  createContext<{
    inputRef: MutableRefObject<HTMLInputElement>;
    inputValue: string;
    onInputValueChange(value: string): void;
    changeActivePicker(picker: SearchPicker): void;
    activePicker: SearchPicker;
    close(): void;
    closeAndResetInput(): void;
    open(): void;
    opened: boolean;
  }>(null);

export const SearchContextProvider: FC = props => {
  const inputRef = useRef<HTMLInputElement>();
  const [opened, setOpened] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const [activePicker, setActivePicker] = useState(actionPicker);

  function changeActivePicker(picker: SearchPicker): void {
    setActivePicker(picker);
    setInputValue('');
  }

  const close = useCallback(() => {
    setOpened(false);
    setActivePicker(actionPicker);
  }, []);

  const closeAndResetInput = useCallback(() => {
    close();
    setInputValue('');
  }, [close]);

  function open(): void {
    setOpened(true);
    inputRef.current?.focus();
  }

  return (
    <SearchContext.Provider
      value={{
        inputRef,
        inputValue,
        onInputValueChange: setInputValue,
        changeActivePicker,
        activePicker,
        close,
        closeAndResetInput,
        opened,
        open,
      }}
      children={props.children}
    />
  );
};

export const useSearchContext = () => {
  const context = useContext(SearchContext);

  if (!context) {
    throw new Error('SearchContext requires SearchContextProvider context.');
  }

  return context;
};
