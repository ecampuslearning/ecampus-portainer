import { useQuery, useQueries } from '@tanstack/react-query';
import { compact, flatMap } from 'lodash';
import { useMemo } from 'react';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withGlobalError } from '@/react-tools/react-query';
import { UserId } from '@/portainer/users/types';

import { Chart, HelmChartsResponse, HelmRepositoriesResponse } from '../types';

/**
 * Get Helm repositories for user
 */
export async function getHelmRepositories(userId: UserId) {
  try {
    const { data } = await axios.get<HelmRepositoriesResponse>(
      `users/${userId}/helm/repositories`
    );
    const repos = compact([
      // compact will remove the global repository if it's empty
      data.GlobalRepository.toLowerCase(),
      ...data.UserRepositories.map((repo) => repo.URL.toLowerCase()),
    ]);
    return [...new Set(repos)];
  } catch (err) {
    throw parseAxiosError(err, 'Unable to retrieve helm repositories for user');
  }
}

async function getChartsFromRepo(repo: string): Promise<Chart[]> {
  try {
    // Construct the URL with required repo parameter
    const response = await axios.get<HelmChartsResponse>('templates/helm', {
      params: { repo },
    });

    return compact(
      Object.values(response.data.entries).map((versions) =>
        versions[0] ? { ...versions[0], repo } : null
      )
    );
  } catch (error) {
    // Ignore errors from chart repositories as some may error but others may not
    return [];
  }
}

/**
 * Hook to fetch all accessible Helm repositories for a user
 *
 * @param userId User ID
 * @returns Query result with list of repository URLs
 */
export function useHelmRepositories(userId: number) {
  return useQuery(
    [userId, 'helm-repositories'],
    () => getHelmRepositories(userId),
    {
      enabled: !!userId,
      ...withGlobalError('Unable to retrieve Helm repositories'),
    }
  );
}

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
