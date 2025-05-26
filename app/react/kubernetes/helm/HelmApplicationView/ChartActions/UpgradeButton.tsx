import { ArrowUp } from 'lucide-react';
import { useRouter } from '@uirouter/react';
import { useState } from 'react';

import { EnvironmentId } from '@/react/portainer/environments/types';
import { notifySuccess } from '@/portainer/services/notifications';
import { semverCompare } from '@/react/common/semver-utils';

import { LoadingButton } from '@@/buttons';
import { InlineLoader } from '@@/InlineLoader';
import { Tooltip } from '@@/Tip/Tooltip';
import { Link } from '@@/Link';

import { HelmRelease } from '../../types';
import {
  useUpdateHelmReleaseMutation,
  UpdateHelmReleasePayload,
} from '../queries/useUpdateHelmReleaseMutation';
import {
  ChartVersion,
  useHelmRepoVersions,
  useHelmRepositories,
} from '../queries/useHelmRepositories';
import { useHelmRelease } from '../queries/useHelmRelease';

import { openUpgradeHelmModal } from './UpgradeHelmModal';

export function UpgradeButton({
  environmentId,
  releaseName,
  namespace,
  release,
  updateRelease,
}: {
  environmentId: EnvironmentId;
  releaseName: string;
  namespace: string;
  release?: HelmRelease;
  updateRelease: (release: HelmRelease) => void;
}) {
  const router = useRouter();
  const updateHelmReleaseMutation = useUpdateHelmReleaseMutation(environmentId);

  const repositoriesQuery = useHelmRepositories();
  const [useCache, setUseCache] = useState(true);
  const helmRepoVersionsQuery = useHelmRepoVersions(
    release?.chart.metadata?.name || '',
    60 * 60 * 1000, // 1 hour
    repositoriesQuery.data,
    useCache
  );
  const versions = helmRepoVersionsQuery.data;

  // Combined loading state
  const isInitialLoading =
    repositoriesQuery.isInitialLoading ||
    helmRepoVersionsQuery.isFetching ||
    helmRepoVersionsQuery.isInitialLoading;
  const isError = repositoriesQuery.isError || helmRepoVersionsQuery.isError;

  const latestVersion = useHelmRelease(environmentId, releaseName, namespace, {
    select: (data) => data.chart.metadata?.version,
  });
  const latestVersionAvailable = versions[0]?.Version ?? '';
  const isNewVersionAvailable = Boolean(
    latestVersion?.data &&
      semverCompare(latestVersionAvailable, latestVersion?.data) === 1
  );

  const editableHelmRelease: UpdateHelmReleasePayload = {
    name: releaseName,
    namespace: namespace || '',
    values: release?.values?.userSuppliedValues,
    chart: release?.chart.metadata?.name || '',
    version: release?.chart.metadata?.version,
  };

  function handleRefreshVersions() {
    if (useCache === false) {
      helmRepoVersionsQuery.refetch();
    } else {
      setUseCache(false);
    }
  }

  return (
    <div className="relative">
      <LoadingButton
        color="secondary"
        data-cy="k8sApp-upgradeHelmChartButton"
        onClick={() => openUpgradeForm(versions, release)}
        disabled={
          versions.length === 0 ||
          isInitialLoading ||
          isError ||
          release?.info?.status?.startsWith('pending')
        }
        loadingText="Upgrading..."
        isLoading={updateHelmReleaseMutation.isLoading}
        icon={ArrowUp}
        size="medium"
      >
        Upgrade
      </LoadingButton>
      {versions.length === 0 && isInitialLoading && (
        <InlineLoader
          size="xs"
          className="absolute -bottom-5 left-0 right-0 whitespace-nowrap"
        >
          Checking for new versions...
        </InlineLoader>
      )}
      {!isInitialLoading && !isError && (
        <span className="absolute flex items-center -bottom-5 left-0 right-0 text-xs text-muted text-center whitespace-nowrap">
          {getStatusMessage(
            versions.length === 0,
            latestVersionAvailable,
            isNewVersionAvailable
          )}
          {versions.length === 0 && (
            <Tooltip
              message={
                <div>
                  Portainer is unable to find any versions for this chart in the
                  repositories saved. Try adding a new repository which contains
                  the chart in the{' '}
                  <Link
                    to="portainer.account"
                    params={{ '#': 'helm-repositories' }}
                    data-cy="user-settings-link"
                  >
                    Helm repositories settings
                  </Link>
                </div>
              }
            />
          )}
          <button
            onClick={handleRefreshVersions}
            className="text-primary hover:text-primary-light cursor-pointer bg-transparent border-0 pl-1 p-0"
            type="button"
          >
            Refresh versions
          </button>
        </span>
      )}
    </div>
  );

  async function openUpgradeForm(
    versions: ChartVersion[],
    release?: HelmRelease
  ) {
    const result = await openUpgradeHelmModal(editableHelmRelease, versions);

    if (result) {
      handleUpgrade(result, release);
    }
  }

  function handleUpgrade(
    payload: UpdateHelmReleasePayload,
    release?: HelmRelease
  ) {
    if (release?.info) {
      const updatedRelease = {
        ...release,
        info: {
          ...release.info,
          status: 'pending-upgrade',
          description: 'Preparing upgrade',
        },
      };
      updateRelease(updatedRelease);
    }
    updateHelmReleaseMutation.mutate(payload, {
      onSuccess: () => {
        notifySuccess('Success', 'Helm chart upgraded successfully');
        // set the revision url param to undefined to refresh the page at the latest revision
        router.stateService.go('kubernetes.helm', {
          namespace,
          name: releaseName,
          revision: undefined,
        });
      },
    });
  }

  function getStatusMessage(
    hasNoAvailableVersions: boolean,
    latestVersionAvailable: string,
    isNewVersionAvailable: boolean
  ): string {
    if (hasNoAvailableVersions) {
      return 'No versions available ';
    }
    if (isNewVersionAvailable) {
      return `New version available (${latestVersionAvailable}) `;
    }
    return '';
  }
}
