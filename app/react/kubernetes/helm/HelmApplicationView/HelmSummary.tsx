import { Badge } from '@/react/components/Badge';

import { Alert } from '@@/Alert';

import { HelmRelease } from '../types';

interface Props {
  release: HelmRelease;
}

export enum DeploymentStatus {
  DEPLOYED = 'deployed',
  FAILED = 'failed',
  PENDING = 'pending-install',
  PENDINGUPGRADE = 'pending-upgrade',
  PENDINGROLLBACK = 'pending-rollback',
  SUPERSEDED = 'superseded',
  UNINSTALLED = 'uninstalled',
  UNINSTALLING = 'uninstalling',
}

export function HelmSummary({ release }: Props) {
  const isSuccess =
    release.info?.status === DeploymentStatus.DEPLOYED ||
    release.info?.status === DeploymentStatus.SUPERSEDED;

  return (
    <div>
      <div className="flex flex-col gap-y-4">
        <div>
          <Badge type={getStatusColor(release.info?.status)}>
            {getText(release.info?.status)}
          </Badge>
        </div>
        <div className="flex flex-wrap gap-2">
          {!!release.namespace && <Badge>Namespace: {release.namespace}</Badge>}
          {!!release.version && <Badge>Revision: #{release.version}</Badge>}
          {!!release.chart?.metadata?.name && (
            <Badge>Chart: {release.chart.metadata.name}</Badge>
          )}
          {!!release.chart?.metadata?.appVersion && (
            <Badge>App version: {release.chart.metadata.appVersion}</Badge>
          )}
          {!!release.chart?.metadata?.version && (
            <Badge>
              Chart version: {release.chart.metadata.name}-
              {release.chart.metadata.version}
            </Badge>
          )}
        </div>
        {!!release.info?.description && !isSuccess && (
          <Alert color={getAlertColor(release.info?.status)}>
            {release.info?.description}
          </Alert>
        )}
      </div>
    </div>
  );
}

function getAlertColor(status?: string) {
  switch (status?.toLowerCase()) {
    case DeploymentStatus.DEPLOYED:
      return 'success';
    case DeploymentStatus.FAILED:
      return 'error';
    case DeploymentStatus.PENDING:
    case DeploymentStatus.PENDINGUPGRADE:
    case DeploymentStatus.PENDINGROLLBACK:
    case DeploymentStatus.UNINSTALLING:
      return 'warn';
    case DeploymentStatus.SUPERSEDED:
    default:
      return 'info';
  }
}

function getStatusColor(status?: string) {
  switch (status?.toLowerCase()) {
    case DeploymentStatus.DEPLOYED:
      return 'success';
    case DeploymentStatus.FAILED:
      return 'danger';
    case DeploymentStatus.PENDING:
    case DeploymentStatus.PENDINGUPGRADE:
    case DeploymentStatus.PENDINGROLLBACK:
    case DeploymentStatus.UNINSTALLING:
      return 'warn';
    case DeploymentStatus.SUPERSEDED:
    default:
      return 'info';
  }
}

function getText(status?: string) {
  switch (status?.toLowerCase()) {
    case DeploymentStatus.DEPLOYED:
      return 'Deployed';
    case DeploymentStatus.FAILED:
      return 'Failed';
    case DeploymentStatus.PENDING:
    case DeploymentStatus.PENDINGUPGRADE:
    case DeploymentStatus.PENDINGROLLBACK:
    case DeploymentStatus.UNINSTALLING:
      return 'Pending';
    case DeploymentStatus.SUPERSEDED:
      return 'Superseded';
    default:
      return 'Unknown';
  }
}
