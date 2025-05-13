import { EventList, Event } from 'kubernetes-types/core/v1';
import { useQuery } from '@tanstack/react-query';

import { EnvironmentId } from '@/react/portainer/environments/types';
import axios from '@/portainer/services/axios';
import { withGlobalError } from '@/react-tools/react-query';

import { parseKubernetesAxiosError } from '../axiosError';

import { queryKeys as environmentQueryKeys } from './query-keys';

type RequestOptions = {
  /** if undefined, events are fetched at the cluster scope */
  namespace?: string;
  params?: {
    /** https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors */
    labelSelector?: string;
    /** https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors */
    fieldSelector?: string;
  };
};

const queryKeys = {
  base: (environmentId: number, { namespace, params }: RequestOptions) => {
    if (namespace) {
      return [
        ...environmentQueryKeys.base(environmentId),
        'events',
        namespace,
        params,
      ] as const;
    }
    return [
      ...environmentQueryKeys.base(environmentId),
      'events',
      params,
    ] as const;
  },
};

async function getEvents(
  environmentId: EnvironmentId,
  options?: RequestOptions
): Promise<Event[]> {
  const { namespace, params } = options ?? {};
  try {
    const { data } = await axios.get<EventList>(
      buildUrl(environmentId, namespace),
      {
        params,
      }
    );
    return data.items;
  } catch (e) {
    throw parseKubernetesAxiosError(e, 'Unable to retrieve events');
  }
}

type QueryOptions<T> = {
  queryOptions?: {
    autoRefreshRate?: number;
    select?: (data: Event[]) => T;
  };
} & RequestOptions;

export function useEvents<T = Event[]>(
  environmentId: EnvironmentId,
  options?: QueryOptions<T>
) {
  const { queryOptions, params, namespace } = options ?? {};
  return useQuery(
    queryKeys.base(environmentId, { params, namespace }),
    () => getEvents(environmentId, { params, namespace }),
    {
      ...withGlobalError('Unable to retrieve events'),
      refetchInterval() {
        return queryOptions?.autoRefreshRate ?? false;
      },
      select: queryOptions?.select,
    }
  );
}

export function useEventWarningsCount(
  environmentId: EnvironmentId,
  namespace?: string
) {
  const resourceEventsQuery = useEvents<number>(environmentId, {
    namespace,
    queryOptions: {
      select: (data) => data.filter((e) => e.type === 'Warning').length,
    },
  });
  return resourceEventsQuery.data || 0;
}

function buildUrl(environmentId: EnvironmentId, namespace?: string) {
  return namespace
    ? `/endpoints/${environmentId}/kubernetes/api/v1/namespaces/${namespace}/events`
    : `/endpoints/${environmentId}/kubernetes/api/v1/events`;
}
