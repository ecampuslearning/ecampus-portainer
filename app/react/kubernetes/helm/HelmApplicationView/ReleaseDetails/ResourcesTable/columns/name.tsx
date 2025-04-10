import { CellContext } from '@tanstack/react-table';

import { Link } from '@@/Link';

import { ResourceRow } from '../types';

import { columnHelper } from './helper';

export const name = columnHelper.accessor((row) => row.name.label, {
  header: 'Name',
  cell: Cell,
  id: 'name',
});

function Cell({ row }: CellContext<ResourceRow, string>) {
  const { name } = row.original;

  if (name.link && name.link.to) {
    return (
      <Link
        to={name.link.to}
        params={name.link.params}
        title={name.label}
        className="w-fit max-w-xs truncate xl:max-w-sm 2xl:max-w-md"
        data-cy={`helm-resource-link-${name.label}`}
      >
        {name.label}
      </Link>
    );
  }

  return name.label;
}
