import { useCurrentStateAndParams } from '@uirouter/react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';

import { Datatable, TableSettingsMenu } from '@@/datatables';
import {
  createPersistedStore,
  refreshableSettings,
  TableSettingsWithRefreshable,
} from '@@/datatables/types';
import { useTableState } from '@@/datatables/useTableState';
import { Widget } from '@@/Widget';
import { TableSettingsMenuAutoRefresh } from '@@/datatables/TableSettingsMenuAutoRefresh';

import { useHelmRelease } from '../../queries/useHelmRelease';

import { columns } from './columns';
import { useResourceRows } from './useResourceRows';

const storageKey = 'helm-resources';

export function createStore(storageKey: string) {
  return createPersistedStore<TableSettingsWithRefreshable>(
    storageKey,
    'name',
    (set) => ({
      ...refreshableSettings(set),
    })
  );
}

const settingsStore = createStore('helm-resources');

export function ResourcesTable() {
  const environmentId = useEnvironmentId();
  const { params } = useCurrentStateAndParams();
  const { name, namespace } = params;

  const tableState = useTableState(settingsStore, storageKey);
  const helmReleaseQuery = useHelmRelease(environmentId, name, namespace, {
    showResources: true,
    refetchInterval: tableState.autoRefreshRate * 1000,
  });
  const rows = useResourceRows(helmReleaseQuery.data?.info?.resources);

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
        renderTableSettings={() => (
          <TableSettingsMenu>
            <TableSettingsMenuAutoRefresh
              value={tableState.autoRefreshRate}
              onChange={(value) => tableState.setAutoRefreshRate(value)}
            />
          </TableSettingsMenu>
        )}
      />
    </Widget>
  );
}
