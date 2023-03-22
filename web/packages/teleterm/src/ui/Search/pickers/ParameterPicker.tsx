import React, { useCallback, useEffect } from 'react';
import { Highlight } from 'shared/components/Highlight';
import { mapAttempt, useAsync } from 'shared/hooks/useAsync';

import { useSearchContext } from '../SearchContext';
import { ParametrizedAction } from '../actions';

import { ResultList } from './ResultList';
import { actionPicker } from './pickers';

interface ParameterPickerProps {
  action: ParametrizedAction;
}

export function ParameterPicker(props: ParameterPickerProps) {
  const { inputValue, close, changeActivePicker } = useSearchContext();
  const [suggestionsAttempt, fetch] = useAsync(
    props.action.parameter.getSuggestions
  );

  useEffect(() => {
    fetch();
  }, [props.action]);

  const attempt = mapAttempt(suggestionsAttempt, suggestions =>
    suggestions.filter(v =>
      v.toLocaleLowerCase().includes(inputValue.toLocaleLowerCase())
    )
  );

  let extraItems: string[] = [];
  if (inputValue) {
    extraItems = [inputValue];
  }

  const onPick = useCallback(
    (item: string) => {
      props.action.perform(item);
      close();
    },
    [close, props.action]
  );

  const onBack = useCallback(() => {
    changeActivePicker(actionPicker);
  }, [changeActivePicker]);

  return (
    <ResultList<string>
      attempt={attempt}
      extraItems={extraItems}
      onPick={onPick}
      onBack={onBack}
      render={item => (
        <Highlight text={item} keywords={[inputValue]}></Highlight>
      )}
    />
  );
}
