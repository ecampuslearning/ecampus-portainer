import { columnHelper } from './helper';

export const resourceType = columnHelper.accessor((row) => row.resourceType, {
  header: 'Resource type',
  id: 'resourceType',
});
