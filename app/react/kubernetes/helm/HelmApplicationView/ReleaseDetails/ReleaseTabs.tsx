import { useState } from 'react';
import { compact } from 'lodash';

import { NavTabs, Option } from '@@/NavTabs';

import { HelmRelease } from '../../types';

import { ManifestDetails } from './ManifestDetails';
import { NotesDetails } from './NotesDetails';
import { ValuesDetails } from './ValuesDetails';
import { ResourcesTable } from './ResourcesTable/ResourcesTable';

type Props = {
  release: HelmRelease;
};

type Tab = 'values' | 'notes' | 'manifest' | 'resources';

function helmTabs(
  release: HelmRelease,
  isUserSupplied: boolean,
  setIsUserSupplied: (isUserSupplied: boolean) => void
): Option<Tab>[] {
  return compact([
    {
      label: 'Resources',
      id: 'resources',
      children: <ResourcesTable resources={release.info?.resources ?? []} />,
    },
    {
      label: 'Values',
      id: 'values',
      children: (
        <ValuesDetails
          values={release.values}
          isUserSupplied={isUserSupplied}
          setIsUserSupplied={setIsUserSupplied}
        />
      ),
    },
    {
      label: 'Manifest',
      id: 'manifest',
      children: <ManifestDetails manifest={release.manifest} />,
    },
    !!release.info?.notes && {
      label: 'Notes',
      id: 'notes',
      children: <NotesDetails notes={release.info.notes} />,
    },
  ]);
}

export function ReleaseTabs({ release }: Props) {
  const [tab, setTab] = useState<Tab>('resources');
  // state is here so that the state isn't lost when the tab changes
  const [isUserSupplied, setIsUserSupplied] = useState(true);

  return (
    <NavTabs<Tab>
      onSelect={setTab}
      selectedId={tab}
      type="pills"
      justified
      options={helmTabs(release, isUserSupplied, setIsUserSupplied)}
    />
  );
}
