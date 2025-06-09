import { useQueries } from '@tanstack/react-query';
import { compact, flatMap } from 'lodash';
import { useMemo } from 'react';

import axios from '@/portainer/services/axios';
import { withGlobalError } from '@/react-tools/react-query';

import { Chart, HelmChartsResponse } from '../types';

/**
 * React hook to fetch helm charts from the provided repositories
 * Charts from each repository are loaded independently, allowing the UI
 * to show charts as they become available instead of waiting for all
 * repositories to load
 *
 * @param userId User ID
 * @param repositories List of repository URLs to fetch charts from
 */
export function useHelmChartList(userId: number, repositories: string[] = []) {
  // Fetch charts from each repository in parallel as separate queries
  const chartQueries = useQueries({
    queries: useMemo(
      () =>
        repositories.map((repo) => ({
          queryKey: [userId, repo, 'helm-charts'],
          queryFn: () => getChartsFromRepo(repo),
          enabled: !!userId && repositories.length > 0,
          // one request takes a long time, so fail early to get feedback to the user faster
          retries: false,
          ...withGlobalError(`Unable to retrieve Helm charts from ${repo}`),
        })),
      [repositories, userId]
    ),
  });

  // Combine the results for easier consumption by components
  const allCharts = useMemo(
    () => flatMap(compact(chartQueries.map((q) => q.data))),
    [chartQueries]
  );

  return {
    // Data from all repositories that have loaded so far
    data: allCharts,
    // Overall loading state
    isInitialLoading: chartQueries.some((q) => q.isInitialLoading),
    // Overall error state
    isError: chartQueries.some((q) => q.isError),
  };
}

async function getChartsFromRepo(repo: string): Promise<Chart[]> {
  try {
    // Construct the URL with required repo parameter
    const response = await axios.get<HelmChartsResponse>('templates/helm', {
      params: { repo },
    });

    return compact(
      Object.values(response.data.entries).map((versions) =>
        versions[0]
          ? {
              ...versions[0],
              repo,
              // versions are within this response too, so we don't need a new query to fetch versions when this is used
              versions: versions.map((v) => v.version),
            }
          : null
      )
    );
  } catch (error) {
    // Ignore errors from chart repositories as some may error but others may not
    return [];
  }
}
