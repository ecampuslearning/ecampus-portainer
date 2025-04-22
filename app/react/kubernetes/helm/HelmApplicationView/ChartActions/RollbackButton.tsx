import { RotateCcw } from 'lucide-react';

import { EnvironmentId } from '@/react/portainer/environments/types';
import { notifySuccess } from '@/portainer/services/notifications';

import { LoadingButton } from '@@/buttons';
import { buildConfirmButton } from '@@/modals/utils';
import { confirm } from '@@/modals/confirm';
import { ModalType } from '@@/modals';

import { useHelmRollbackMutation } from '../queries/useHelmRollbackMutation';

type Props = {
  latestRevision: number;
  environmentId: EnvironmentId;
  releaseName: string;
  namespace?: string;
};

export function RollbackButton({
  latestRevision,
  environmentId,
  releaseName,
  namespace,
}: Props) {
  // the selectedRevision can be a prop when selecting a revision is implemented
  const selectedRevision = latestRevision ? latestRevision - 1 : undefined;

  const rollbackMutation = useHelmRollbackMutation(environmentId);

  return (
    <LoadingButton
      onClick={handleClick}
      isLoading={rollbackMutation.isLoading}
      loadingText="Rolling back..."
      data-cy="rollback-button"
      icon={RotateCcw}
      color="default"
      size="medium"
    >
      Rollback to #{selectedRevision}
    </LoadingButton>
  );

  async function handleClick() {
    const confirmed = await confirm({
      title: 'Are you sure?',
      modalType: ModalType.Warn,
      confirmButton: buildConfirmButton('Rollback'),
      message: `Rolling back will restore the application to revision #${selectedRevision}, which will cause service interruption. Do you wish to continue?`,
    });
    if (!confirmed) {
      return;
    }

    rollbackMutation.mutate(
      {
        releaseName,
        params: { namespace, revision: selectedRevision },
      },
      {
        onSuccess: () => {
          notifySuccess(
            'Success',
            `Application rolled back to revision #${selectedRevision} successfully.`
          );
        },
      }
    );
  }
}
