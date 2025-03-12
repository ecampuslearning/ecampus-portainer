import {
  Loading,
  Widget,
  WidgetBody,
  WidgetTitle,
} from '@/react/components/Widget';
import helm from '@/assets/ico/vendor/helm.svg?c';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';

import { Alert } from '@@/Alert';

import { useHelmRelease } from './queries/useHelmRelease';

interface HelmDetailsWidgetProps {
  name: string;
  namespace: string;
}

export function HelmDetailsWidget({ name, namespace }: HelmDetailsWidgetProps) {
  const environmentId = useEnvironmentId();

  const {
    data: release,
    isInitialLoading,
    isError,
  } = useHelmRelease(environmentId, name, namespace);

  return (
    <Widget>
      <WidgetTitle icon={helm} title="Release" />
      <WidgetBody>
        {isInitialLoading && <Loading />}

        {isError && (
          <Alert
            color="error"
            title="Failed to load Helm application details"
          />
        )}

        {!isInitialLoading && !isError && release && (
          <table className="table">
            <tbody>
              <tr>
                <td className="!border-none w-40">Name</td>
                <td
                  className="!border-none min-w-[140px]"
                  data-cy="k8sAppDetail-appName"
                >
                  {release.name}
                </td>
              </tr>
              <tr>
                <td className="!border-t">Chart</td>
                <td className="!border-t">{release.chart}</td>
              </tr>
              <tr>
                <td>App version</td>
                <td>{release.app_version}</td>
              </tr>
            </tbody>
          </table>
        )}
      </WidgetBody>
    </Widget>
  );
}
