import { useCurrentStateAndParams } from '@uirouter/react';

import helm from '@/assets/ico/vendor/helm.svg?c';
import { PageHeader } from '@/react/components/PageHeader';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { Authorized } from '@/react/hooks/useUser';

import { WidgetTitle, WidgetBody, Widget, Loading } from '@@/Widget';
import { Card } from '@@/Card';
import { Alert } from '@@/Alert';

import { HelmRelease } from '../types';

import { HelmSummary } from './HelmSummary';
import { ReleaseTabs } from './ReleaseDetails/ReleaseTabs';
import { useHelmRelease } from './queries/useHelmRelease';
import { ChartActions } from './ChartActions/ChartActions';

export function HelmApplicationView() {
  const environmentId = useEnvironmentId();
  const { params } = useCurrentStateAndParams();
  const { name, namespace } = params;

  const helmReleaseQuery = useHelmRelease(environmentId, name, namespace, {
    showResources: true,
  });

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
          <Widget>
            {name && (
              <WidgetTitle icon={helm} title={name}>
                <Authorized authorizations="K8sApplicationsW">
                  <ChartActions
                    environmentId={environmentId}
                    releaseName={name}
                    namespace={namespace}
                    currentRevision={helmReleaseQuery.data?.version}
                  />
                </Authorized>
              </WidgetTitle>
            )}
            <WidgetBody>
              <HelmDetails
                isLoading={helmReleaseQuery.isInitialLoading}
                isError={helmReleaseQuery.isError}
                release={helmReleaseQuery.data}
              />
            </WidgetBody>
          </Widget>
        </div>
      </div>
    </>
  );
}

type HelmDetailsProps = {
  isLoading: boolean;
  isError: boolean;
  release: HelmRelease | undefined;
};

function HelmDetails({ isLoading, isError, release: data }: HelmDetailsProps) {
  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return (
      <Alert color="error" title="Failed to load Helm application details" />
    );
  }

  if (!data) {
    return <Alert color="error" title="No Helm application details found" />;
  }

  return (
    <>
      <HelmSummary release={data} />
      <div className="my-6 h-[1px] w-full bg-gray-5 th-dark:bg-gray-7 th-highcontrast:bg-white" />
      <Card className="bg-inherit">
        <ReleaseTabs release={data} />
      </Card>
    </>
  );
}
