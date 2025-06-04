import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MutationOptions } from '@tanstack/react-query';
import { vi } from 'vitest';

import { withTestQueryProvider } from '@/react/test-utils/withTestQuery';
import { withUserProvider } from '@/react/test-utils/withUserProvider';
import { withTestRouter } from '@/react/test-utils/withRouter';
import { UserViewModel } from '@/portainer/models/user';

import { Chart } from '../types';

import { HelmTemplatesSelectedItem } from './HelmTemplatesSelectedItem';

const mockMutate = vi.fn();
const mockNotifySuccess = vi.fn();

// Mock dependencies
vi.mock('@/portainer/services/notifications', () => ({
  notifySuccess: (title: string, text: string) =>
    mockNotifySuccess(title, text),
}));

vi.mock('./queries/useHelmChartValues', () => ({
  useHelmChartValues: vi.fn().mockReturnValue({
    data: { values: 'test-values' },
    isLoading: false,
  }),
}));

vi.mock('./queries/useHelmChartInstall', () => ({
  useHelmChartInstall: vi.fn().mockReturnValue({
    mutate: (params: Record<string, string>, options?: MutationOptions) =>
      mockMutate(params, options),
    isLoading: false,
  }),
}));

vi.mock('@/react/hooks/useAnalytics', () => ({
  useAnalytics: vi.fn().mockReturnValue({
    trackEvent: vi.fn(),
  }),
}));

// Sample test data
const mockChart: Chart = {
  name: 'test-chart',
  description: 'Test Chart Description',
  repo: 'https://example.com',
  icon: 'test-icon-url',
  annotations: {
    category: 'database',
  },
};

const clearHelmChartMock = vi.fn();
const mockRouterStateService = {
  go: vi.fn(),
};

function renderComponent({
  selectedChart = mockChart,
  clearHelmChart = clearHelmChartMock,
  namespace = 'test-namespace',
  name = 'test-name',
} = {}) {
  const user = new UserViewModel({ Username: 'user' });

  const Wrapped = withTestQueryProvider(
    withUserProvider(
      withTestRouter(() => (
        <HelmTemplatesSelectedItem
          selectedChart={selectedChart}
          clearHelmChart={clearHelmChart}
          namespace={namespace}
          name={name}
        />
      )),
      user
    )
  );

  return {
    ...render(<Wrapped />),
    user,
    mockRouterStateService,
  };
}

describe('HelmTemplatesSelectedItem', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should display selected chart information', () => {
    renderComponent();

    // Check for chart details
    expect(screen.getByText('test-chart')).toBeInTheDocument();
    expect(screen.getByText('Test Chart Description')).toBeInTheDocument();
    expect(screen.getByText('Clear selection')).toBeInTheDocument();
    expect(screen.getByText('https://example.com')).toBeInTheDocument();
  });

  it('should toggle custom values editor', async () => {
    renderComponent();
    const user = userEvent.setup();

    // Verify editor is visible by default
    expect(screen.getByTestId('helm-app-creation-editor')).toBeInTheDocument();

    // Now hide the editor
    await user.click(await screen.findByText('Custom values'));

    // Editor should be hidden
    expect(
      screen.queryByTestId('helm-app-creation-editor')
    ).not.toBeInTheDocument();
  });

  it('should install helm chart and navigate when install button is clicked', async () => {
    const user = userEvent.setup();
    renderComponent();

    // Click install button
    await user.click(screen.getByText('Install'));

    // Check mutate was called with correct values
    expect(mockMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        Name: 'test-name',
        Repo: 'https://example.com',
        Chart: 'test-chart',
        Values: 'test-values',
        Namespace: 'test-namespace',
      }),
      expect.objectContaining({ onSuccess: expect.any(Function) })
    );
  });

  it('should disable install button when namespace or name is undefined', () => {
    renderComponent({ namespace: '' });
    expect(screen.getByText('Install')).toBeDisabled();
  });
});
