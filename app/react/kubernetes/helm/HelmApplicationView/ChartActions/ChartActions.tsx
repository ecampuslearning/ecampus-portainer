import { EnvironmentId } from '@/react/portainer/environments/types';
import { useAuthorizations } from '@/react/hooks/useUser';

import { UninstallButton } from './UninstallButton';

export function ChartActions({
  environmentId,
  releaseName,
  namespace,
}: {
  environmentId: EnvironmentId;
  releaseName: string;
  namespace?: string;
}) {
  const { authorized } = useAuthorizations('K8sApplicationsW');

  if (!authorized) {
    return null;
  }

  return (
    <div className="inline-flex gap-x-2">
      <UninstallButton
        environmentId={environmentId}
        releaseName={releaseName}
        namespace={namespace}
      />
    </div>
  );
}
