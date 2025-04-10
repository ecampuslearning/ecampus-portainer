import { render, screen } from '@testing-library/react';
import { HttpResponse } from 'msw';

import { withTestQueryProvider } from '@/react/test-utils/withTestQuery';
import { server, http } from '@/setup-tests/server';
import { withTestRouter } from '@/react/test-utils/withRouter';
import { UserViewModel } from '@/portainer/models/user';
import { withUserProvider } from '@/react/test-utils/withUserProvider';
import { mockCodeMirror } from '@/setup-tests/mock-codemirror';

import { HelmApplicationView } from './HelmApplicationView';

// Mock the necessary hooks and dependencies
const mockUseCurrentStateAndParams = vi.fn();
const mockUseEnvironmentId = vi.fn();

vi.mock('@uirouter/react', async (importOriginal: () => Promise<object>) => ({
  ...(await importOriginal()),
  useCurrentStateAndParams: () => mockUseCurrentStateAndParams(),
}));

vi.mock('@/react/hooks/useEnvironmentId', () => ({
  useEnvironmentId: () => mockUseEnvironmentId(),
}));

mockCodeMirror();

const minimalHelmRelease = {
  name: 'test-release',
  version: '1',
  namespace: 'default',
  chart: {
    metadata: {
      name: 'test-chart',
      // appVersion: '1.0.0', // can be missing for a minimal release
      version: '2.2.2',
    },
  },
  info: {
    status: 'deployed',
    // notes: 'This is a test note', // can be missing for a minimal release
  },
  manifest: 'This is a test manifest',
};

const helmReleaseWithAdditionalDetails = {
  ...minimalHelmRelease,
  info: {
    ...minimalHelmRelease.info,
    notes: 'This is a test note',
  },
  chart: {
    ...minimalHelmRelease.chart,
    metadata: {
      ...minimalHelmRelease.chart.metadata,
      appVersion: '1.0.0',
    },
  },
};

function renderComponent() {
  const user = new UserViewModel({ Username: 'user' });
  const Wrapped = withTestQueryProvider(
    withUserProvider(withTestRouter(HelmApplicationView), user)
  );
  return render(<Wrapped />);
}

describe('HelmApplicationView', () => {
  beforeEach(() => {
    // Set up default mock values
    mockUseEnvironmentId.mockReturnValue(3);
    mockUseCurrentStateAndParams.mockReturnValue({
      params: {
        name: 'test-release',
        namespace: 'default',
      },
    });
  });

  it('should display helm release details for minimal release when data is loaded', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});

    server.use(
      http.get('/api/endpoints/3/kubernetes/helm/test-release', () =>
        HttpResponse.json(minimalHelmRelease)
      )
    );

    const { findByText, findAllByText } = renderComponent();

    // Check for the page header
    expect(await findByText('Helm details')).toBeInTheDocument();

    // Check for the badge content
    expect(await findByText(/Namespace/)).toBeInTheDocument();
    expect(await findByText(/Chart version:/)).toBeInTheDocument();
    expect(await findByText(/Chart:/)).toBeInTheDocument();
    expect(await findByText(/Revision/)).toBeInTheDocument();

    // Check for the actual values
    expect(await findAllByText(/test-release/)).toHaveLength(2); // title and badge
    expect(await findAllByText(/test-chart/)).toHaveLength(2);

    // There shouldn't be a notes tab when there are no notes
    expect(screen.queryByText(/Notes/)).not.toBeInTheDocument();

    // There shouldn't be an app version badge when it's missing
    expect(screen.queryByText(/App version/)).not.toBeInTheDocument();

    // Ensure there are no console errors
    // eslint-disable-next-line no-console
    expect(console.error).not.toHaveBeenCalled();

    // Restore console.error
    vi.spyOn(console, 'error').mockRestore();
  });

  it('should display error message when API request fails', async () => {
    // Mock API failure
    server.use(
      http.get('/api/endpoints/3/kubernetes/helm/test-release', () =>
        HttpResponse.error()
      )
    );

    // Mock console.error to prevent test output pollution
    vi.spyOn(console, 'error').mockImplementation(() => {});

    renderComponent();

    // Wait for the error message to appear
    expect(
      await screen.findByText('Failed to load Helm application details')
    ).toBeInTheDocument();

    // Restore console.error
    vi.spyOn(console, 'error').mockRestore();
  });

  it('should display additional details when available in helm release', async () => {
    server.use(
      http.get('/api/endpoints/3/kubernetes/helm/test-release', () =>
        HttpResponse.json(helmReleaseWithAdditionalDetails)
      )
    );

    const { findByText } = renderComponent();

    // Check for the notes tab when notes are available
    expect(await findByText(/Notes/)).toBeInTheDocument();

    // Check for the app version badge when it's available
    expect(await findByText(/App version/)).toBeInTheDocument();
    expect(await findByText('1.0.0', { exact: false })).toBeInTheDocument();
  });
});
