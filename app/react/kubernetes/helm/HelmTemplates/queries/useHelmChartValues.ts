import { useQuery } from '@tanstack/react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withGlobalError } from '@/react-tools/react-query';

import { Chart } from '../../types';

async function getHelmChartValues(chart: string, repo: string) {
  try {
    const response = await axios.get<string>(`/templates/helm/values`, {
      params: {
        repo,
        chart,
      },
    });
    return response.data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to get Helm chart values');
  }
}

export function useHelmChartValues(chart: Chart) {
  return useQuery({
    queryKey: ['helm-chart-values', chart.repo, chart.name],
    queryFn: () => getHelmChartValues(chart.name, chart.repo),
    enabled: !!chart.name,
    select: (data) => ({
      values: data,
    }),
    ...withGlobalError('Unable to get Helm chart values'),
  });
}
