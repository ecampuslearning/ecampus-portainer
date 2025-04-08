import { useMutation } from '@tanstack/react-query';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import {
  queryClient,
  withGlobalError,
  withInvalidate,
} from '@/react-tools/react-query';
import { queryKeys } from '@/react/kubernetes/applications/queries/query-keys';

import { InstallChartPayload } from '../../types';

async function installHelmChart(
  payload: InstallChartPayload,
  environmentId: EnvironmentId
) {
  try {
    const response = await axios.post(
      `endpoints/${environmentId}/kubernetes/helm`,
      payload
    );
    return response.data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Installation error');
  }
}

export function useHelmChartInstall() {
  const environmentId = useEnvironmentId();

  return useMutation(
    (values: InstallChartPayload) => installHelmChart(values, environmentId),
    {
      ...withGlobalError('Unable to install Helm chart'),
      ...withInvalidate(queryClient, [queryKeys.applications(environmentId)]),
    }
  );
}
