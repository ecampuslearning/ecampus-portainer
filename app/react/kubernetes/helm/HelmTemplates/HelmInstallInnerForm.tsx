import { Form, useFormikContext } from 'formik';
import { useMemo } from 'react';

import { FormActions } from '@@/form-components/FormActions';
import { FormControl } from '@@/form-components/FormControl';
import { Option, PortainerSelect } from '@@/form-components/PortainerSelect';
import { FormSection } from '@@/form-components/FormSection';

import { ChartVersion } from '../queries/useHelmRepositories';
import { Chart } from '../types';
import { useHelmChartValues } from '../queries/useHelmChartValues';
import { HelmValuesInput } from '../components/HelmValuesInput';

import { HelmInstallFormValues } from './types';

type Props = {
  selectedChart: Chart;
  namespace?: string;
  name?: string;
  versionOptions: Option<ChartVersion>[];
  isLoadingVersions: boolean;
};

export function HelmInstallInnerForm({
  selectedChart,
  namespace,
  name,
  versionOptions,
  isLoadingVersions,
}: Props) {
  const { values, setFieldValue, isSubmitting } =
    useFormikContext<HelmInstallFormValues>();

  const chartValuesRefQuery = useHelmChartValues({
    chart: selectedChart.name,
    repo: selectedChart.repo,
    version: values?.version,
  });

  const selectedVersion = useMemo(
    () =>
      versionOptions.find((v) => v.value.Version === values.version)?.value ??
      versionOptions[0]?.value,
    [versionOptions, values.version]
  );

  return (
    <Form className="form-horizontal">
      <div className="form-group !m-0">
        <FormSection title="Configuration" className="mt-4">
          <FormControl
            label="Version"
            inputId="version-input"
            isLoading={isLoadingVersions}
            loadingText="Loading versions..."
          >
            <PortainerSelect<ChartVersion>
              value={selectedVersion}
              options={versionOptions}
              onChange={(version) => {
                if (version) {
                  setFieldValue('version', version.Version);
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

      <FormActions
        submitLabel="Install"
        loadingText="Installing Helm chart"
        isLoading={isSubmitting}
        isValid={!!namespace && !!name}
        data-cy="helm-install"
      />
    </Form>
  );
}
