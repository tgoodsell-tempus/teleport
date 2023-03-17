import styled from 'styled-components';

import Table from 'design/DataTable';

export const StyledTable = styled(Table)`
  & > tbody > tr > td {
    vertical-align: middle;
    padding: 8px;
  }
` as typeof Table;
