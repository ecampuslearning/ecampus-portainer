import { Datatable } from '@@/datatables';
import { createPersistedStore } from '@@/datatables/types';
import { useTableState } from '@@/datatables/useTableState';
import { Widget } from '@@/Widget';

import { GenericResource } from '../../../types';

import { columns } from './columns';
import { useResourceRows } from './useResourceRows';

type Props = {
  resources: GenericResource[];
};

const storageKey = 'helm-resources';
const settingsStore = createPersistedStore(storageKey, 'resourceType');

export function ResourcesTable({ resources }: Props) {
  const tableState = useTableState(settingsStore, storageKey);
  const rows = useResourceRows(resources);

  return (
    <Widget>
      <Datatable
        // no widget to avoid extra padding from app/react/components/datatables/TableContainer.tsx
        noWidget
        dataset={rows}
        columns={columns}
        includeSearch
        settingsManager={tableState}
        emptyContentLabel="No resources found"
        disableSelect
        getRowId={(row) => row.id}
        data-cy="helm-resources-datatable"
      />
    </Widget>
  );
}
