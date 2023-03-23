import React from 'react';

import { ParametrizedAction } from '../actions';

import { ActionPicker } from './ActionPicker';
import { ParameterPicker } from './ParameterPicker';

// TODO: Refactor this to a const.
export const getActionPicker = (): SearchPicker => {
  return {
    picker: <ActionPicker />,
    placeholder: 'Search for something',
  };
};
export const getParameterPicker = (
  parametrizedAction: ParametrizedAction
): SearchPicker => {
  return {
    picker: <ParameterPicker action={parametrizedAction} />,
    placeholder: parametrizedAction.parameter.placeholder,
  };
};

interface SearchPicker {
  picker: React.ReactElement;
  placeholder: string;
}
