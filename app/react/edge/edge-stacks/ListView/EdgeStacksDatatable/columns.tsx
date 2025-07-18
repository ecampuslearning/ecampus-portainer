import { createColumnHelper } from '@tanstack/react-table';
import _ from 'lodash';

import { isoDateFromTimestamp } from '@/portainer/filters/filters';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import { GitCommitLink } from '@/react/portainer/gitops/GitCommitLink';

import { buildNameColumnFromObject } from '@@/datatables/buildNameColumn';
import { Link } from '@@/Link';
import { Tooltip } from '@@/Tip/Tooltip';

import { StatusType } from '../../types';

import { EdgeStackStatus } from './EdgeStacksStatus';
import { DecoratedEdgeStack } from './types';
import { DeploymentCounter } from './DeploymentCounter';

const columnHelper = createColumnHelper<DecoratedEdgeStack>();

export const columns = _.compact([
  buildNameColumnFromObject<DecoratedEdgeStack>({
    nameKey: 'Name',
    path: 'edge.stacks.edit',
    dataCy: 'edge-stacks-name',
    idParam: 'stackId',
  }),
  columnHelper.accessor(
    (item) =>
      item.StatusSummary?.AggregatedStatus?.[StatusType.Acknowledged] || 0,
    {
      header: 'Acknowledged',
      enableSorting: false,
      enableHiding: false,
      cell: ({ getValue, row }) => (
        <DeploymentCounter
          count={getValue()}
          type={StatusType.Acknowledged}
          total={row.original.NumDeployments}
        />
      ),
      meta: {
        className: '[&>*]:justify-center',
      },
    }
  ),
  isBE &&
    columnHelper.accessor(
      (item) =>
        item.StatusSummary?.AggregatedStatus?.[StatusType.ImagesPulled] || 0,
      {
        header: 'Images pre-pulled',
        cell: ({ getValue, row: { original: item } }) => {
          if (!item.PrePullImage) {
            return <div className="text-center">-</div>;
          }

          return (
            <DeploymentCounter
              count={getValue()}
              type={StatusType.ImagesPulled}
              total={item.NumDeployments}
            />
          );
        },
        enableSorting: false,
        enableHiding: false,
        meta: {
          className: '[&>*]:justify-center',
        },
      }
    ),
  columnHelper.accessor(
    (item) =>
      item.StatusSummary?.AggregatedStatus?.[StatusType.DeploymentReceived] ||
      0,
    {
      header: 'Deployments received',
      cell: ({ getValue, row }) => (
        <DeploymentCounter
          count={getValue()}
          type={StatusType.Running}
          total={row.original.NumDeployments}
        />
      ),
      enableSorting: false,
      enableHiding: false,
      meta: {
        className: '[&>*]:justify-center',
      },
    }
  ),
  columnHelper.accessor(
    (item) => item.StatusSummary?.AggregatedStatus?.[StatusType.Error] || 0,
    {
      header: 'Deployments failed',
      cell: ({ getValue, row }) => {
        const count = getValue();

        return (
          <div className="flex items-center gap-2">
            <DeploymentCounter
              count={count}
              type={StatusType.Error}
              total={row.original.NumDeployments}
            />
            {count > 0 && (
              <Link
                className="hover:no-underline"
                to="edge.stacks.edit"
                params={{
                  stackId: row.original.Id,
                  tab: 'environments',
                  status: StatusType.Error,
                }}
                data-cy={`edge-stacks-error-${row.original.Id}`}
              >
                ({count}/{row.original.NumDeployments})
              </Link>
            )}
          </div>
        );
      },
      enableSorting: false,
      enableHiding: false,
      meta: {
        className: '[&>*]:justify-center',
      },
    }
  ),
  columnHelper.accessor('Status', {
    header: StatusHeader,
    cell: ({ row }) => (
      <div className="w-full text-center">
        <EdgeStackStatus edgeStack={row.original} />
      </div>
    ),
    enableSorting: false,
    enableHiding: false,
    meta: {
      className: '[&>*]:justify-center',
    },
  }),
  columnHelper.accessor('CreationDate', {
    header: 'Creation Date',
    cell: ({ getValue }) => isoDateFromTimestamp(getValue()),
    enableHiding: false,
  }),
  isBE &&
    columnHelper.accessor(
      (item) =>
        item.GitConfig ? item.GitConfig.ConfigHash : item.StackFileVersion,
      {
        header: 'Target Version',
        enableSorting: false,
        cell: ({ row: { original: item } }) => {
          if (item.GitConfig) {
            return (
              <div className="text-center">
                <GitCommitLink
                  baseURL={item.GitConfig.URL}
                  commitHash={item.GitConfig.ConfigHash}
                />
              </div>
            );
          }

          return <div className="text-center">{item.StackFileVersion}</div>;
        },
        meta: {
          className: '[&>*]:justify-center',
        },
      }
    ),
]);

function StatusHeader() {
  return (
    <>
      Status
      <Tooltip
        position="top"
        message={
          <>
            <div>
              The status feature for the Edge stack is only available for Edge
              Agent versions 2.19.0 and above.
            </div>
            <div>
              To access the status of your edge stack, it is essential to
              upgrade your Edge Agent to a corresponding version that is
              compatible with your Portainer server.
            </div>
          </>
        }
      />
    </>
  );
}
