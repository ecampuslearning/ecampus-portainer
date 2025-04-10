import { CellContext } from '@tanstack/react-table';

import { StatusBadge } from '@@/StatusBadge';

import { ResourceRow } from '../types';

import { columnHelper } from './helper';

export const status = columnHelper.accessor((row) => row.status.label, {
  header: 'Status',
  id: 'status',
  cell: Cell,
});

function Cell({ row }: CellContext<ResourceRow, string>) {
  const { status } = row.original;
  return <StatusBadge color={status.type}>{status.label}</StatusBadge>;
}
