import { useQuery } from '@tanstack/react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withGlobalError } from '@/react-tools/react-query';

type Params = {
  chart: string;
  repo: string;
  version?: string;
};

async function getHelmChartValues(params: Params) {
  try {
    const response = await axios.get<string>(`/templates/helm/values`, {
      params,
    });
    return response.data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to get Helm chart values');
  }
}

export function useHelmChartValues(params: Params) {
  return useQuery({
    queryKey: ['helm-chart-values', params.repo, params.chart, params.version],
    queryFn: () => getHelmChartValues(params),
    enabled: !!params.chart && !!params.repo,
    select: (data) => ({
      values: data,
    }),
    staleTime: 60 * 1000 * 20, // 60 minutes, because values are not expected to change often
    ...withGlobalError('Unable to get Helm chart values'),
  });
}
