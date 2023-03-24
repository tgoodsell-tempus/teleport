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

import React from 'react';
import styled from 'styled-components';

interface UserIconProps {
  letter: string;
}

export function UserIcon(props: UserIconProps) {
  return <Circle>{props.letter.toLocaleUpperCase()}</Circle>;
}

const Circle = styled.span`
  border-radius: 50%;
  color: ${props => props.theme.colors.light};
  background: ${props => props.theme.colors.brand.main};
  height: 24px;
  width: 24px;
  display: flex;
  flex-shrink: 0;
  justify-content: center;
  align-items: center;
  overflow: hidden;
`;
