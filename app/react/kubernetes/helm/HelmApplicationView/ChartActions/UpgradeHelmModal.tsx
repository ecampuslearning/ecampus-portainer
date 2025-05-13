import { useState } from 'react';
import { ArrowUp } from 'lucide-react';

import { withReactQuery } from '@/react-tools/withReactQuery';
import { withCurrentUser } from '@/react-tools/withCurrentUser';

import { Modal, OnSubmit, openModal } from '@@/modals';
import { Button } from '@@/buttons';
import { Option, PortainerSelect } from '@@/form-components/PortainerSelect';
import { Input } from '@@/form-components/Input';
import { CodeEditor } from '@@/CodeEditor';
import { FormControl } from '@@/form-components/FormControl';
import { WidgetTitle } from '@@/Widget';

import { UpdateHelmReleasePayload } from '../queries/useUpdateHelmReleaseMutation';
import { ChartVersion } from '../queries/useHelmRepositories';

interface Props {
  onSubmit: OnSubmit<UpdateHelmReleasePayload>;
  values: UpdateHelmReleasePayload;
  versions: ChartVersion[];
}

export function UpgradeHelmModal({ values, versions, onSubmit }: Props) {
  const versionOptions: Option<ChartVersion>[] = versions.map((version) => {
    const isCurrentVersion = version.Version === values.version;
    const label = `${version.Repo}@${version.Version}${
      isCurrentVersion ? ' (current)' : ''
    }`;
    return {
      label,
      value: version,
    };
  });
  const defaultVersion =
    versionOptions.find((v) => v.value.Version === values.version)?.value ||
    versionOptions[0]?.value;
  const [version, setVersion] = useState<ChartVersion>(defaultVersion);
  const [userValues, setUserValues] = useState<string>(values.values || '');

  return (
    <Modal
      onDismiss={() => onSubmit()}
      size="lg"
      className="flex flex-col h-[80vh] px-0"
      aria-label="upgrade-helm"
    >
      <Modal.Header
        title={<WidgetTitle className="px-5" title="Upgrade" icon={ArrowUp} />}
      />
      <div className="flex-1 overflow-y-auto px-5">
        <Modal.Body>
          <FormControl label="Version" inputId="version-input" size="vertical">
            <PortainerSelect<ChartVersion>
              value={version}
              options={versionOptions}
              onChange={(version) => {
                if (version) {
                  setVersion(version);
                }
              }}
              data-cy="helm-version-input"
            />
          </FormControl>
          <FormControl
            label="Release name"
            inputId="release-name-input"
            size="vertical"
          >
            <Input
              id="release-name-input"
              value={values.name}
              readOnly
              disabled
              data-cy="helm-release-name-input"
            />
          </FormControl>
          <FormControl
            label="Namespace"
            inputId="namespace-input"
            size="vertical"
          >
            <Input
              id="namespace-input"
              value={values.namespace}
              readOnly
              disabled
              data-cy="helm-namespace-input"
            />
          </FormControl>
          <FormControl
            label="User-defined values"
            inputId="user-values-editor"
            size="vertical"
          >
            <CodeEditor
              id="user-values-editor"
              value={userValues}
              onChange={(value) => setUserValues(value)}
              height="50vh"
              type="yaml"
              data-cy="helm-user-values-editor"
              placeholder="Define or paste the content of your values yaml file here"
            />
          </FormControl>
        </Modal.Body>
      </div>
      <div className="px-5 border-solid border-0 border-t border-gray-5 th-dark:border-gray-7 th-highcontrast:border-white">
        <Modal.Footer>
          <Button
            onClick={() => onSubmit()}
            color="secondary"
            key="cancel-button"
            size="medium"
            data-cy="cancel-button-cy"
          >
            Cancel
          </Button>
          <Button
            onClick={() =>
              onSubmit({
                name: values.name,
                values: userValues,
                namespace: values.namespace,
                chart: values.chart,
                repo: version.Repo,
                version: version.Version,
              })
            }
            color="primary"
            key="update-button"
            size="medium"
            data-cy="update-button-cy"
          >
            Upgrade
          </Button>
        </Modal.Footer>
      </div>
    </Modal>
  );
}

export async function openUpgradeHelmModal(
  values: UpdateHelmReleasePayload,
  versions: ChartVersion[]
) {
  return openModal(withReactQuery(withCurrentUser(UpgradeHelmModal)), {
    values,
    versions,
  });
}
