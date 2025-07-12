import { Form, useFormikContext } from 'formik';
import { useMemo } from 'react';

import { FormControl } from '@@/form-components/FormControl';
import { Option, PortainerSelect } from '@@/form-components/PortainerSelect';
import { FormSection } from '@@/form-components/FormSection';
import { LoadingButton } from '@@/buttons';

import { Chart } from '../types';
import { useHelmChartValues } from '../queries/useHelmChartValues';
import { HelmValuesInput } from '../components/HelmValuesInput';
import { ChartVersion } from '../queries/useHelmRepoVersions';

import { HelmInstallFormValues } from './types';

type Props = {
  selectedChart: Chart;
  namespace?: string;
  name?: string;
  versionOptions: Option<ChartVersion>[];
  isVersionsLoading: boolean;
  isRepoAvailable: boolean;
};

export function HelmInstallInnerForm({
  selectedChart,
  namespace,
  name,
  versionOptions,
  isVersionsLoading,
  isRepoAvailable,
}: Props) {
  const { values, setFieldValue, isSubmitting } =
    useFormikContext<HelmInstallFormValues>();

  const selectedVersion: ChartVersion | undefined = useMemo(
    () =>
      versionOptions.find(
        (v) =>
          v.value.Version === values.version &&
          v.value.Repo === selectedChart.repo
      )?.value ?? versionOptions[0]?.value,
    [versionOptions, values.version, selectedChart.repo]
  );

  const repoParams = {
    repo: selectedChart.repo,
  };
  // use isLatestVersionFetched to cache the latest version, to avoid duplicate fetches
  const isLatestVersionFetched =
    // if no version is selected, the latest version gets fetched
    !versionOptions.length ||
    // otherwise check if the selected version is the latest version
    (selectedVersion?.Version === versionOptions[0]?.value.Version &&
      selectedVersion?.Repo === versionOptions[0]?.value.Repo);
  const chartValuesRefQuery = useHelmChartValues(
    {
      chart: selectedChart.name,
      version: values?.version,
      ...repoParams,
    },
    isLatestVersionFetched
  );

  return (
    <Form className="form-horizontal">
      <div className="form-group !m-0">
        <FormSection title="Configuration" className="mt-4">
          <FormControl
            label="Version"
            inputId="version-input"
            isLoading={isVersionsLoading}
            loadingText="Loading versions..."
          >
            <PortainerSelect<ChartVersion>
              value={selectedVersion}
              options={versionOptions}
              noOptionsMessage={() => 'No versions found'}
              placeholder="Select a version"
              onChange={(version) => {
                if (version) {
                  setFieldValue('version', version.Version);
                  setFieldValue('repo', version.Repo);
                }
              }}
              data-cy="helm-version-input"
            />
          </FormControl>
          <HelmValuesInput
            values={values.values}
            setValues={(values) => setFieldValue('values', values)}
            valuesRef={chartValuesRefQuery.data?.values ?? ''}
            isValuesRefLoading={chartValuesRefQuery.isInitialLoading}
          />
        </FormSection>
      </div>

      <LoadingButton
        className="!ml-0"
        loadingText="Installing Helm chart"
        isLoading={isSubmitting}
        disabled={!namespace || !name || !isRepoAvailable}
        data-cy="helm-install"
      >
        Install
      </LoadingButton>
    </Form>
  );
}
