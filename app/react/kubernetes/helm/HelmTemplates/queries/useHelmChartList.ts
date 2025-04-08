import { useQuery } from '@tanstack/react-query';
import { compact } from 'lodash';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withGlobalError } from '@/react-tools/react-query';

import {
  Chart,
  HelmChartsResponse,
  HelmRepositoriesResponse,
} from '../../types';

async function getHelmRepositories(userId: number): Promise<string[]> {
  try {
    const response = await axios.get<HelmRepositoriesResponse>(
      `users/${userId}/helm/repositories`
    );
    const { GlobalRepository, UserRepositories } = response.data;

    // Extract URLs from user repositories
    const userHelmReposUrls = UserRepositories.map((repo) => repo.URL);

    // Combine global and user repositories, remove duplicates and empty values
    const uniqueHelmRepos = [
      ...new Set([GlobalRepository, ...userHelmReposUrls]),
    ]
      .map((url) => url.toLowerCase())
      .filter((url) => url);

    return uniqueHelmRepos;
  } catch (err) {
    throw parseAxiosError(err, 'Failed to fetch Helm repositories');
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

async function getCharts(userId: number): Promise<Chart[]> {
  try {
    // First, get all the helm repositories
    const repos = await getHelmRepositories(userId);

    // Then fetch charts from each repository in parallel
    const chartsPromises = repos.map((repo) => getChartsFromRepo(repo));
    const chartsArrays = await Promise.all(chartsPromises);

    // Flatten the arrays of charts into a single array
    return chartsArrays.flat();
  } catch (err) {
    throw parseAxiosError(err, 'Failed to fetch Helm charts');
  }
}

/**
 * React hook to fetch helm charts from all accessible repositories
 * @param userId User ID
 */
export function useHelmChartList(userId: number) {
  return useQuery([userId, 'helm-charts'], () => getCharts(userId), {
    enabled: !!userId,
    ...withGlobalError('Unable to retrieve Helm charts'),
  });
}
