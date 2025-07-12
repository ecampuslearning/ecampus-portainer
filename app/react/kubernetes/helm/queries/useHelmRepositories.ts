import { useQuery } from '@tanstack/react-query';
import { compact } from 'lodash';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { UserId } from '@/portainer/users/types';
import { withGlobalError } from '@/react-tools/react-query';
import { useCurrentUser } from '@/react/hooks/useUser';
import { Option } from '@/react/components/form-components/PortainerSelect';

import { HelmRegistriesResponse } from '../types';
import { RepoValue } from '../components/HelmRegistrySelect';

/**
 * Hook to fetch all Helm registries for the current user
 */
export function useUserHelmRepositories<T = string[]>({
  select,
}: {
  select?: (registries: string[]) => T;
} = {}) {
  const { user } = useCurrentUser();
  return useQuery(
    ['helm', 'registries'],
    async () => getUserHelmRepositories(user.Id),
    {
      enabled: !!user.Id,
      select,
      ...withGlobalError('Unable to retrieve helm registries'),
    }
  );
}

export function useHelmRepoOptions() {
  return useUserHelmRepositories({
    select: (registries) => {
      const repoOptions = registries
        .map<Option<RepoValue>>((registry) => ({
          label: registry,
          value: {
            repoUrl: registry,
            isOCI: false,
            name: registry,
          },
        }))
        .sort((a, b) => a.label.localeCompare(b.label));
      return [
        {
          label: 'Helm Repositories',
          options: repoOptions,
        },
        {
          label: 'OCI Registries',
          options: [
            {
              label:
                'Installing from an OCI registry is a Portainer Business Feature',
              value: {},
              disabled: true,
            },
          ],
        },
      ];
    },
  });
}

/**
 * Get Helm repositories for user
 */
async function getUserHelmRepositories(userId: UserId) {
  try {
    const { data } = await axios.get<HelmRegistriesResponse>(
      `users/${userId}/helm/repositories`
    );
    // compact will remove the global repository if it's empty
    const repos = compact([
      data.GlobalRepository.toLowerCase(),
      ...data.UserRepositories.map((repo) => repo.URL.toLowerCase()),
    ]);
    return [...new Set(repos)];
  } catch (err) {
    throw parseAxiosError(err, 'Unable to retrieve helm repositories for user');
  }
}
