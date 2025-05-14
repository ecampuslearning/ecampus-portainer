import { useQueryClient, useMutation } from '@tanstack/react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withGlobalError, withInvalidate } from '@/react-tools/react-query';
import { queryKeys as applicationsQueryKeys } from '@/react/kubernetes/applications/queries/query-keys';
import { EnvironmentId } from '@/react/portainer/environments/types';

import { HelmRelease } from '../../types';

export interface UpdateHelmReleasePayload {
  namespace: string;
  values?: string;
  repo?: string;
  name: string;
  chart: string;
  version?: string;
  atomic?: boolean;
}
export function useUpdateHelmReleaseMutation(environmentId: EnvironmentId) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: UpdateHelmReleasePayload) =>
      updateHelmRelease(environmentId, payload),
    ...withInvalidate(queryClient, [
      [environmentId, 'helm', 'releases'],
      applicationsQueryKeys.applications(environmentId),
    ]),
    ...withGlobalError('Unable to uninstall helm application'),
  });
}

async function updateHelmRelease(
  environmentId: EnvironmentId,
  payload: UpdateHelmReleasePayload
) {
  try {
    const { data } = await axios.post<HelmRelease>(
      `endpoints/${environmentId}/kubernetes/helm`,
      payload
    );
    return data;
  } catch (err) {
    throw parseAxiosError(err, 'Unable to update helm release');
  }
}
