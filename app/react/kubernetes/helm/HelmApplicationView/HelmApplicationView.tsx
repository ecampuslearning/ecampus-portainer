import { useCurrentStateAndParams } from '@uirouter/react';

import { PageHeader } from '@/react/components/PageHeader';

import { HelmDetailsWidget } from './HelmDetailsWidget';

export function HelmApplicationView() {
  const { params } = useCurrentStateAndParams();

  const { name, namespace } = params;

  return (
    <>
      <PageHeader
        title="Helm details"
        breadcrumbs={[
          { label: 'Applications', link: 'kubernetes.applications' },
          name,
        ]}
        reload
      />

      <div className="row">
        <div className="col-sm-12">
          <HelmDetailsWidget name={name} namespace={namespace} />
        </div>
      </div>
    </>
  );
}
