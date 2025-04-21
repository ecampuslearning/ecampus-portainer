import { useCurrentStateAndParams } from '@uirouter/react';

import helm from '@/assets/ico/vendor/helm.svg?c';
import { PageHeader } from '@/react/components/PageHeader';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { EnvironmentId } from '@/react/portainer/environments/types';

import { WidgetTitle, WidgetBody, Widget, Loading } from '@@/Widget';
import { Card } from '@@/Card';
import { Alert } from '@@/Alert';

import { HelmSummary } from './HelmSummary';
import { ReleaseTabs } from './ReleaseDetails/ReleaseTabs';
import { useHelmRelease } from './queries/useHelmRelease';
import { ChartActions } from './ChartActions/ChartActions';

export function HelmApplicationView() {
  const environmentId = useEnvironmentId();
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
          <Widget>
            {name && (
              <WidgetTitle icon={helm} title={name}>
                <ChartActions
                  environmentId={environmentId}
                  releaseName={name}
                  namespace={namespace}
                />
              </WidgetTitle>
            )}
            <WidgetBody>
              <HelmDetails
                name={name}
                namespace={namespace}
                environmentId={environmentId}
              />
            </WidgetBody>
          </Widget>
        </div>
      </div>
    </>
  );
}

type HelmDetailsProps = {
  name: string;
  namespace: string;
  environmentId: EnvironmentId;
};

function HelmDetails({ name, namespace, environmentId }: HelmDetailsProps) {
  const {
    data: release,
    isInitialLoading,
    isError,
  } = useHelmRelease(environmentId, name, namespace, {
    showResources: true,
  });

  if (isInitialLoading) {
    return <Loading />;
  }

  if (isError) {
    return (
      <Alert color="error" title="Failed to load Helm application details" />
    );
  }

  if (!release) {
    return <Alert color="error" title="No Helm application details found" />;
  }

  return (
    <>
      <HelmSummary release={release} />
      <div className="my-6 h-[1px] w-full bg-gray-5 th-dark:bg-gray-7 th-highcontrast:bg-white" />
      <Card className="bg-inherit">
        <ReleaseTabs release={release} />
      </Card>
    </>
  );
}
