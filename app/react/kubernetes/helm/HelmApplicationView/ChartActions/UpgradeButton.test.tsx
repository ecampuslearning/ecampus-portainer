import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { withTestQueryProvider } from '@/react/test-utils/withTestQuery';
import { withTestRouter } from '@/react/test-utils/withRouter';

import {
  useHelmRepoVersions,
  ChartVersion,
} from '../queries/useHelmRepositories';
import { HelmRelease } from '../../types';

import { openUpgradeHelmModal } from './UpgradeHelmModal';
import { UpgradeButton } from './UpgradeButton';

// Mock the upgrade modal function
vi.mock('./UpgradeHelmModal', () => ({
  openUpgradeHelmModal: vi.fn(() => Promise.resolve(undefined)),
}));

// Mock the notifications service
vi.mock('@/portainer/services/notifications', () => ({
  notifySuccess: vi.fn(),
}));

// Mock the useHelmRepoVersions and useHelmRepositories hooks
vi.mock('../queries/useHelmRepositories', () => ({
  useHelmRepoVersions: vi.fn(() => ({
    data: [
      { Version: '1.0.0', Repo: 'stable' },
      { Version: '1.1.0', Repo: 'stable' },
    ],
    isInitialLoading: false,
    isError: false,
  })),
  useHelmRepositories: vi.fn(() => ({
    data: ['repo1', 'repo2'],
    isInitialLoading: false,
    isError: false,
  })),
}));

function renderButton(props = {}) {
  const defaultProps = {
    environmentId: 1,
    releaseName: 'test-release',
    namespace: 'default',
    release: {
      name: 'test-release',
      chart: {
        metadata: {
          name: 'test-chart',
          version: '1.0.0',
        },
      },
      values: {
        userSuppliedValues: '{}',
      },
      manifest: '',
    } as HelmRelease,
    updateRelease: vi.fn(),
    ...props,
  };

  const Wrapped = withTestQueryProvider(withTestRouter(UpgradeButton));
  return render(<Wrapped {...defaultProps} />);
}

describe('UpgradeButton', () => {
  test('should display the upgrade button', () => {
    renderButton();

    const button = screen.getByRole('button', { name: /Upgrade/i });
    expect(button).toBeInTheDocument();
  });

  test('should be disabled when no versions are available', () => {
    const data: ChartVersion[] = [];
    vi.mocked(useHelmRepoVersions).mockReturnValue({
      data,
      isInitialLoading: false,
      isError: false,
    });

    renderButton();

    const button = screen.getByRole('button', { name: /Upgrade/i });
    expect(button).toBeDisabled();
  });

  test('should show loading state when checking for versions', () => {
    vi.mocked(useHelmRepoVersions).mockReturnValue({
      data: [],
      isInitialLoading: true,
      isError: false,
    });

    renderButton();

    expect(
      screen.getByText('Checking for new versions...')
    ).toBeInTheDocument();
  });

  test('should show "No versions available" when no versions are found', () => {
    const data: ChartVersion[] = [];
    vi.mocked(useHelmRepoVersions).mockReturnValue({
      data,
      isInitialLoading: false,
      isError: false,
    });

    renderButton();

    expect(screen.getByText('No versions available')).toBeInTheDocument();
  });

  test('should open upgrade modal when clicked', async () => {
    const user = userEvent.setup();
    const mockRelease = {
      name: 'test-release',
      chart: {
        metadata: {
          name: 'test-chart',
          version: '1.0.0',
        },
      },
      values: {
        userSuppliedValues: '{}',
      },
      manifest: '',
    } as HelmRelease;

    vi.mocked(useHelmRepoVersions).mockReturnValue({
      data: [
        { Version: '1.0.0', Repo: 'stable' },
        { Version: '1.1.0', Repo: 'stable' },
      ],
      isInitialLoading: false,
      isError: false,
    });

    renderButton({ release: mockRelease });

    const button = screen.getByRole('button', { name: /Upgrade/i });
    await user.click(button);

    await waitFor(() => {
      expect(openUpgradeHelmModal).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'test-release',
          chart: 'test-chart',
          namespace: 'default',
          values: '{}',
          version: '1.0.0',
        }),
        expect.arrayContaining([
          { Version: '1.0.0', Repo: 'stable' },
          { Version: '1.1.0', Repo: 'stable' },
        ])
      );
    });
  });

  test('should not execute the upgrade if modal is cancelled', async () => {
    const mockUpdateRelease = vi.fn();
    vi.mocked(openUpgradeHelmModal).mockResolvedValueOnce(undefined);

    const user = userEvent.setup();
    renderButton({ updateRelease: mockUpdateRelease });

    const button = screen.getByRole('button', { name: /Upgrade/i });
    await user.click(button);

    await waitFor(() => {
      expect(openUpgradeHelmModal).toHaveBeenCalled();
    });

    expect(mockUpdateRelease).not.toHaveBeenCalled();
  });
});
