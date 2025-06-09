import { useQuery } from '@tanstack/react-query';
import { compact } from 'lodash';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { UserId } from '@/portainer/users/types';
import { withGlobalError } from '@/react-tools/react-query';
import { useCurrentUser } from '@/react/hooks/useUser';

import { HelmRegistriesResponse } from '../types';

/**
 * Hook to fetch all Helm registries for the current user
 */
export function useHelmRegistries() {
  const { user } = useCurrentUser();
  return useQuery(
    ['helm', 'registries'],
    async () => getHelmRegistries(user.Id),
    {
      enabled: !!user.Id,
      ...withGlobalError('Unable to retrieve helm registries'),
    }
  );
}

/**
 * Get Helm registries for user
 */
async function getHelmRegistries(userId: UserId) {
  try {
    const { data } = await axios.get<HelmRegistriesResponse>(
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
