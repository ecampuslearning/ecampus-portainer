import { render, screen, waitFor } from '@testing-library/react';
import { HttpResponse } from 'msw';
import { Event, EventList } from 'kubernetes-types/core/v1';

import { server, http } from '@/setup-tests/server';
import { withTestQueryProvider } from '@/react/test-utils/withTestQuery';
import { withTestRouter } from '@/react/test-utils/withRouter';
import { UserViewModel } from '@/portainer/models/user';
import { withUserProvider } from '@/react/test-utils/withUserProvider';
import { mockLocalizeDate } from '@/setup-tests/mock-localizeDate';

import { GenericResource } from '../../types';

import {
  HelmEventsDatatable,
  filterRelatedEvents,
} from './HelmEventsDatatable';

const mockUseEnvironmentId = vi.fn();
mockLocalizeDate();

vi.mock('@/react/hooks/useEnvironmentId', () => ({
  useEnvironmentId: () => mockUseEnvironmentId(),
}));

const testResources: GenericResource[] = [
  {
    kind: 'Deployment',
    status: {
      healthSummary: {
        status: 'Healthy',
        reason: 'Running',
        message: 'All replicas are ready',
      },
    },
    metadata: {
      name: 'test-deployment',
      namespace: 'default',
      uid: 'test-deployment-uid',
    },
  },
  {
    kind: 'Service',
    status: {
      healthSummary: {
        status: 'Healthy',
        reason: 'Available',
        message: 'Service is available',
      },
    },
    metadata: {
      name: 'test-service',
      namespace: 'default',
      uid: 'test-service-uid',
    },
  },
];

const mockEventsResponse: EventList = {
  kind: 'EventList',
  apiVersion: 'v1',
  metadata: {
    resourceVersion: '12345',
  },
  items: [
    {
      metadata: {
        name: 'test-deployment-123456',
        namespace: 'default',
        uid: 'event-uid-1',
        resourceVersion: '1000',
        creationTimestamp: '2023-01-01T00:00:00Z',
      },
      involvedObject: {
        kind: 'Deployment',
        namespace: 'default',
        name: 'test-deployment',
        uid: 'test-deployment-uid',
        apiVersion: 'apps/v1',
        resourceVersion: '2000',
      },
      reason: 'ScalingReplicaSet',
      message: 'Scaled up replica set test-deployment-abc123 to 1',
      source: {
        component: 'deployment-controller',
      },
      firstTimestamp: '2023-01-01T00:00:00Z',
      lastTimestamp: '2023-01-01T00:00:00Z',
      count: 1,
      type: 'Normal',
      reportingComponent: 'deployment-controller',
      reportingInstance: '',
    },
    {
      metadata: {
        name: 'test-service-123456',
        namespace: 'default',
        uid: 'event-uid-2',
        resourceVersion: '1001',
        creationTimestamp: '2023-01-01T00:00:00Z',
      },
      involvedObject: {
        kind: 'Service',
        namespace: 'default',
        name: 'test-service',
        uid: 'test-service-uid',
        apiVersion: 'v1',
        resourceVersion: '2001',
      },
      reason: 'CreatedLoadBalancer',
      message: 'Created load balancer',
      source: {
        component: 'service-controller',
      },
      firstTimestamp: '2023-01-01T00:00:00Z',
      lastTimestamp: '2023-01-01T00:00:00Z',
      count: 1,
      type: 'Normal',
      reportingComponent: 'service-controller',
      reportingInstance: '',
    },
  ],
};

const mixedEventsResponse: EventList = {
  kind: 'EventList',
  apiVersion: 'v1',
  metadata: {
    resourceVersion: '12345',
  },
  items: [
    {
      metadata: {
        name: 'test-deployment-123456',
        namespace: 'default',
        uid: 'event-uid-1',
        resourceVersion: '1000',
        creationTimestamp: '2023-01-01T00:00:00Z',
      },
      involvedObject: {
        kind: 'Deployment',
        namespace: 'default',
        name: 'test-deployment',
        uid: 'test-deployment-uid', // This matches a resource UID
        apiVersion: 'apps/v1',
        resourceVersion: '2000',
      },
      reason: 'ScalingReplicaSet',
      message: 'Scaled up replica set test-deployment-abc123 to 1',
      source: {
        component: 'deployment-controller',
      },
      firstTimestamp: '2023-01-01T00:00:00Z',
      lastTimestamp: '2023-01-01T00:00:00Z',
      count: 1,
      type: 'Normal',
      reportingComponent: 'deployment-controller',
      reportingInstance: '',
    },
    {
      metadata: {
        name: 'unrelated-pod-123456',
        namespace: 'default',
        uid: 'event-uid-3',
        resourceVersion: '1002',
        creationTimestamp: '2023-01-01T00:00:00Z',
      },
      involvedObject: {
        kind: 'Pod',
        namespace: 'default',
        name: 'unrelated-pod',
        uid: 'unrelated-pod-uid', // This does NOT match any resource UIDs
        apiVersion: 'v1',
        resourceVersion: '2002',
      },
      reason: 'Scheduled',
      message: 'Successfully assigned unrelated-pod to node',
      source: {
        component: 'default-scheduler',
      },
      firstTimestamp: '2023-01-01T00:00:00Z',
      lastTimestamp: '2023-01-01T00:00:00Z',
      count: 1,
      reportingComponent: 'scheduler',
      reportingInstance: '',
    },
  ],
};

function renderComponent() {
  const user = new UserViewModel({ Username: 'user' });
  mockUseEnvironmentId.mockReturnValue(3);

  const HelmEventsDatatableWithProviders = withTestQueryProvider(
    withUserProvider(withTestRouter(HelmEventsDatatable), user)
  );

  return render(
    <HelmEventsDatatableWithProviders
      namespace="default"
      releaseResources={testResources}
    />
  );
}

describe('HelmEventsDatatable', () => {
  beforeEach(() => {
    server.use(
      http.get(
        '/api/endpoints/3/kubernetes/api/v1/namespaces/default/events',
        () => HttpResponse.json(mockEventsResponse)
      )
    );
  });

  it('should render events datatable with correct title', async () => {
    renderComponent();

    await waitFor(() => {
      expect(
        screen.getByText(
          'Only events for resources currently in the cluster will be displayed.'
        )
      ).toBeInTheDocument();
    });

    expect(screen.getByRole('table')).toBeInTheDocument();
  });

  it('should correctly filter related events using the filterRelatedEvents function', () => {
    const filteredEvents = filterRelatedEvents(
      mixedEventsResponse.items as Event[],
      testResources
    );

    expect(filteredEvents.length).toBe(1);
    expect(filteredEvents[0].involvedObject.uid).toBe('test-deployment-uid');

    const unrelatedEvents = filteredEvents.filter(
      (e) => e.involvedObject.uid === 'unrelated-pod-uid'
    );
    expect(unrelatedEvents.length).toBe(0);
  });
});
