import React from 'react';

import { ParametrizedAction } from '../actions';

import { ActionPicker } from './ActionPicker';
import { ParameterPicker } from './ParameterPicker';

export const actionPicker: SearchPicker = {
  picker: <ActionPicker />,
  placeholder: 'Search for something',
};
export const getParameterPicker = (
  parametrizedAction: ParametrizedAction
): SearchPicker => {
  return {
    picker: <ParameterPicker action={parametrizedAction} />,
    placeholder: parametrizedAction.parameter.placeholder,
  };
};

export interface SearchPicker {
  picker: React.ReactElement;
  placeholder: string;
}
