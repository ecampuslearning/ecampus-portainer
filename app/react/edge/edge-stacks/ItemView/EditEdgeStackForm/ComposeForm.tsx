import { useFormikContext } from 'formik';

import { useDockerComposeSchema } from '@/react/hooks/useDockerComposeSchema/useDockerComposeSchema';

import { TextTip } from '@@/Tip/TextTip';
import { WebEditorForm } from '@@/WebEditorForm';

import { DeploymentType } from '../../types';

import { FormValues } from './types';

export function ComposeForm({
  handleContentChange,
  hasKubeEndpoint,
  handleVersionChange,
  versionOptions,
}: {
  hasKubeEndpoint: boolean;
  handleContentChange: (type: DeploymentType, content: string) => void;
  handleVersionChange: (newVersion: number) => void;
  versionOptions: number[] | undefined;
}) {
  const { errors, values } = useFormikContext<FormValues>();
  const { data: dockerComposeSchema } = useDockerComposeSchema();

  return (
    <>
      {hasKubeEndpoint && (
        <TextTip>
          <p>
            Portainer no longer supports{' '}
            <a
              href="https://docs.docker.com/compose/compose-file/"
              target="_blank"
              rel="noreferrer"
            >
              docker-compose
            </a>{' '}
            format manifests for Kubernetes deployments, and we have removed the{' '}
            <a href="https://kompose.io/" target="_blank" rel="noreferrer">
              Kompose
            </a>{' '}
            conversion tool which enables this. The reason for this is because
            Kompose now poses a security risk, since it has a number of Common
            Vulnerabilities and Exposures (CVEs).
          </p>
          <p>
            Unfortunately, while the Kompose project has a maintainer and is
            part of the CNCF, it is not being actively maintained. Releases are
            very infrequent and new pull requests to the project (including ones
            we&apos;ve submitted) are taking months to be merged, with new CVEs
            arising in the meantime.
          </p>
          <p>
            We advise installing your own instance of Kompose in a sandbox
            environment, performing conversions of your Docker Compose files to
            Kubernetes manifests and using those manifests to set up
            applications.
          </p>
        </TextTip>
      )}

      <WebEditorForm
        data-cy="compose-editor"
        value={values.content}
        type="yaml"
        schema={dockerComposeSchema}
        id="compose-editor"
        textTip="Define or paste the content of your docker compose file here"
        onChange={(value) => handleContentChange(DeploymentType.Compose, value)}
        error={errors.content}
        readonly={hasKubeEndpoint}
        versions={versionOptions}
        onVersionChange={handleVersionChange}
      >
        <div>
          You can get more information about Compose file format in the{' '}
          <a
            href="https://docs.docker.com/compose/compose-file/"
            target="_blank"
            rel="noreferrer"
          >
            official documentation
          </a>
          .
        </div>
      </WebEditorForm>
    </>
  );
}
