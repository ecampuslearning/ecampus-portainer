import { useMemo } from 'react';
import { Shuffle } from 'lucide-react';
import { useRouter } from '@uirouter/react';
import clsx from 'clsx';
import { Row } from '@tanstack/react-table';

import {
  Namespaces,
  PortainerNamespace,
} from '@/react/kubernetes/namespaces/types';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { Authorized, useAuthorizations } from '@/react/hooks/useUser';
import { notifyError, notifySuccess } from '@/portainer/services/notifications';
import { pluralize } from '@/portainer/helpers/strings';
import { DefaultDatatableSettings } from '@/react/kubernetes/datatables/DefaultDatatableSettings';
import { SystemResourceDescription } from '@/react/kubernetes/datatables/SystemResourceDescription';
import { useNamespacesQuery } from '@/react/kubernetes/namespaces/queries/useNamespacesQuery';
import { CreateFromManifestButton } from '@/react/kubernetes/components/CreateFromManifestButton';
import { createStore } from '@/react/kubernetes/datatables/default-kube-datatable-store';

import { Datatable, Table, TableSettingsMenu } from '@@/datatables';
import { useTableState } from '@@/datatables/useTableState';
import { DeleteButton } from '@@/buttons/DeleteButton';

import { useMutationDeleteServices, useClusterServices } from '../../service';
import { Service } from '../../types';

import { columns } from './columns';
import { ServiceRowData } from './types';

const storageKey = 'k8sServicesDatatable';
const settingsStore = createStore(storageKey);

export function ServicesDatatable() {
  const tableState = useTableState(settingsStore, storageKey);
  const environmentId = useEnvironmentId();
  const { data: namespaces, ...namespacesQuery } = useNamespacesQuery(
    environmentId,
    {
      select: (namespacesArray) =>
        namespacesArray?.reduce<Record<string, PortainerNamespace>>(
          (acc, namespace) => {
            acc[namespace.Name] = namespace;
            return acc;
          },
          {}
        ),
    }
  );
  const { authorized: canWrite } = useAuthorizations(['K8sServiceW']);
  const { authorized: canAccessSystemResources } = useAuthorizations(
    'K8sAccessSystemNamespaces'
  );
  const { data: services, ...servicesQuery } = useClusterServices(
    environmentId,
    {
      autoRefreshRate: tableState.autoRefreshRate * 1000,
      withApplications: true,
      select: (services) =>
        services?.filter(
          (service) =>
            (canAccessSystemResources && tableState.showSystemResources) ||
            !namespaces?.[service.Namespace]?.IsSystem
        ),
    }
  );

  const servicesWithIsSystem = useServicesRowData(services || [], namespaces);

  return (
    <Datatable
      dataset={servicesWithIsSystem || []}
      columns={columns}
      settingsManager={tableState}
      isLoading={
        servicesQuery.isInitialLoading || namespacesQuery.isInitialLoading
      }
      emptyContentLabel="No services found"
      title="Services"
      titleIcon={Shuffle}
      getRowId={(row) => row.UID}
      isRowSelectable={(row) => !namespaces?.[row.original.Namespace]?.IsSystem}
      disableSelect={!canWrite}
      renderTableActions={(selectedRows) => (
        <TableActions selectedItems={selectedRows} />
      )}
      renderTableSettings={() => (
        <TableSettingsMenu>
          <DefaultDatatableSettings settings={tableState} />
        </TableSettingsMenu>
      )}
      description={
        <SystemResourceDescription
          showSystemResources={
            tableState.showSystemResources || !canAccessSystemResources
          }
        />
      }
      renderRow={servicesRenderRow}
      data-cy="k8s-services-datatable"
    />
  );
}

// useServicesRowData appends the `isSystem` property to the service data
function useServicesRowData(
  services: Service[],
  namespaces?: Namespaces
): ServiceRowData[] {
  return useMemo(
    () =>
      services.map((service) => ({
        ...service,
        IsSystem: namespaces
          ? namespaces?.[service.Namespace]?.IsSystem
          : false,
      })),
    [services, namespaces]
  );
}

// needed to apply custom styling to the row cells and not globally.
// required in the AC's for this ticket.
function servicesRenderRow(
  row: Row<ServiceRowData>,
  highlightedItemId?: string
) {
  return (
    <Table.Row<ServiceRowData>
      cells={row.getVisibleCells()}
      className={clsx('[&>td]:!py-4 [&>td]:!align-top', {
        active: highlightedItemId === row.id,
      })}
    />
  );
}

interface SelectedService {
  Namespace: string;
  Name: string;
}

type TableActionsProps = {
  selectedItems: ServiceRowData[];
};

function TableActions({ selectedItems }: TableActionsProps) {
  const environmentId = useEnvironmentId();
  const deleteServicesMutation = useMutationDeleteServices(environmentId);
  const router = useRouter();

  return (
    <Authorized authorizations="K8sServicesW">
      <DeleteButton
        disabled={selectedItems.length === 0}
        onConfirmed={() => handleRemoveClick(selectedItems)}
        confirmMessage={
          <>
            <p>{`Are you sure you want to remove the selected ${pluralize(
              selectedItems.length,
              'service'
            )}?`}</p>
            <ul className="pl-6">
              {selectedItems.map((s, index) => (
                <li key={index}>
                  {s.Namespace}/{s.Name}
                </li>
              ))}
            </ul>
          </>
        }
        data-cy="k8s-remove-services-button"
      />

      <CreateFromManifestButton data-cy="k8s-create-service-button" />
    </Authorized>
  );

  async function handleRemoveClick(services: SelectedService[]) {
    const payload: Record<string, string[]> = {};
    services.forEach((service) => {
      payload[service.Namespace] = payload[service.Namespace] || [];
      payload[service.Namespace].push(service.Name);
    });

    deleteServicesMutation.mutate(
      { environmentId, data: payload },
      {
        onSuccess: () => {
          notifySuccess(
            'Services successfully removed',
            services.map((s) => `${s.Namespace}/${s.Name}`).join(', ')
          );
          router.stateService.reload();
        },
        onError: (error) => {
          notifyError(
            'Unable to delete service(s)',
            error as Error,
            services.map((s) => `${s.Namespace}/${s.Name}`).join(', ')
          );
        },
      }
    );
  }
}
