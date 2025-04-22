import { useQuery } from '@tanstack/react-query';

import { EnvironmentId } from '@/react/portainer/environments/types';
import { withGlobalError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';

import { HelmRelease } from '../../types';

/**
 * React hook to fetch a specific Helm release
 */
export function useHelmRelease<T = HelmRelease>(
  environmentId: EnvironmentId,
  name: string,
  namespace: string,
  options: {
    select?: (data: HelmRelease) => T;
    showResources?: boolean;
    refetchInterval?: number;
  } = {}
) {
  const { select, showResources, refetchInterval } = options;
  return useQuery(
    [environmentId, 'helm', 'releases', namespace, name, options.showResources],
    () =>
      getHelmRelease(environmentId, name, {
        namespace,
        showResources,
      }),
    {
      enabled: !!environmentId && !!name && !!namespace,
      ...withGlobalError('Unable to retrieve helm application details'),
      select,
      refetchInterval,
    }
  );
}

/**
 * Get a specific Helm release
 */
async function getHelmRelease(
  environmentId: EnvironmentId,
  name: string,
  params: {
    namespace: string;
    showResources?: boolean;
  }
) {
  try {
    const { data } = await axios.get<HelmRelease>(
      `endpoints/${environmentId}/kubernetes/helm/${name}`,
      {
        params,
      }
    );
    return data;
  } catch (err) {
    throw parseAxiosError(err, 'Unable to retrieve helm application details');
  }
}
