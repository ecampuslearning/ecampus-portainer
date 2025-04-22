import { EnvironmentId } from '@/react/portainer/environments/types';

import { RollbackButton } from './RollbackButton';
import { UninstallButton } from './UninstallButton';

export function ChartActions({
  environmentId,
  releaseName,
  namespace,
  currentRevision,
}: {
  environmentId: EnvironmentId;
  releaseName: string;
  namespace?: string;
  currentRevision?: number;
}) {
  const hasPreviousRevision = currentRevision && currentRevision >= 2;

  return (
    <div className="inline-flex gap-x-2">
      <UninstallButton
        environmentId={environmentId}
        releaseName={releaseName}
        namespace={namespace}
      />
      {hasPreviousRevision && (
        <RollbackButton
          latestRevision={currentRevision}
          environmentId={environmentId}
          releaseName={releaseName}
          namespace={namespace}
        />
      )}
    </div>
  );
}
