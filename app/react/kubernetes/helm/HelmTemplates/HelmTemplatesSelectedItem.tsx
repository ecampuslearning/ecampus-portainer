import { useRef } from 'react';
import { X } from 'lucide-react';
import { Form, Formik, FormikProps } from 'formik';
import { useRouter } from '@uirouter/react';

import { notifySuccess } from '@/portainer/services/notifications';
import { useAnalytics } from '@/react/hooks/useAnalytics';
import { useCanExit } from '@/react/hooks/useCanExit';

import { Widget } from '@@/Widget';
import { Button } from '@@/buttons/Button';
import { FallbackImage } from '@@/FallbackImage';
import Svg from '@@/Svg';
import { Icon } from '@@/Icon';
import { WebEditorForm } from '@@/WebEditorForm';
import { confirmGenericDiscard } from '@@/modals/confirm';
import { FormSection } from '@@/form-components/FormSection';
import { InlineLoader } from '@@/InlineLoader';
import { FormActions } from '@@/form-components/FormActions';

import { Chart } from '../types';

import { useHelmChartValues } from './queries/useHelmChartValues';
import { HelmIcon } from './HelmIcon';
import { useHelmChartInstall } from './queries/useHelmChartInstall';

type Props = {
  selectedChart: Chart;
  clearHelmChart: () => void;
  namespace?: string;
  name?: string;
};

type FormValues = {
  values: string;
};

const emptyValues: FormValues = {
  values: '',
};

export function HelmTemplatesSelectedItem({
  selectedChart,
  clearHelmChart,
  namespace,
  name,
}: Props) {
  const router = useRouter();
  const analytics = useAnalytics();

  const { mutate: installHelmChart, isLoading: isInstalling } =
    useHelmChartInstall();
  const { data: initialValues, isLoading: loadingValues } =
    useHelmChartValues(selectedChart);

  const formikRef = useRef<FormikProps<FormValues>>(null);
  useCanExit(() => !formikRef.current?.dirty || confirmGenericDiscard());

  function handleSubmit(values: FormValues) {
    if (!name || !namespace) {
      // Theoretically this should never happen and is mainly to keep typescript happy
      return;
    }

    installHelmChart(
      {
        Name: name,
        Repo: selectedChart.repo,
        Chart: selectedChart.name,
        Values: values.values,
        Namespace: namespace,
      },
      {
        onSuccess() {
          analytics.trackEvent('kubernetes-helm-install', {
            category: 'kubernetes',
            metadata: {
              'chart-name': selectedChart.name,
            },
          });
          notifySuccess('Success', 'Helm chart successfully installed');

          // Reset the form so page can be navigated away from without getting "Are you sure?"
          formikRef.current?.resetForm();
          router.stateService.go('kubernetes.applications');
        },
      }
    );
  }

  return (
    <>
      <Widget>
        <div className="flex">
          <div className="basis-3/4 rounded-[8px] m-2 bg-gray-4 th-highcontrast:bg-black th-highcontrast:text-white th-dark:bg-gray-iron-10 th-dark:text-white">
            <div className="vertical-center p-5">
              <FallbackImage
                src={selectedChart.icon}
                fallbackIcon={HelmIcon}
                className="h-16 w-16"
              />
              <div className="col-sm-12">
                <div className="flex justify-between">
                  <span>
                    <span className="text-2xl font-bold">
                      {selectedChart.name}
                    </span>
                    <span className="space-left pr-2 text-xs">
                      <span className="vertical-center">
                        <Svg icon="helm" className="icon icon-primary" />
                      </span>{' '}
                      <span>Helm</span>
                    </span>
                  </span>
                </div>
                <div className="text-muted text-xs">
                  {selectedChart.description}
                </div>
              </div>
            </div>
          </div>
          <div className="basis-1/4">
            <div className="h-full w-full vertical-center justify-end pr-5">
              <Button
                color="link"
                className="!text-gray-8 hover:no-underline th-highcontrast:!text-white th-dark:!text-white"
                onClick={clearHelmChart}
                data-cy="clear-selection"
              >
                Clear selection
                <Icon icon={X} className="ml-1" />
              </Button>
            </div>
          </div>
        </div>
      </Widget>
      <Formik
        innerRef={formikRef}
        initialValues={initialValues ?? emptyValues}
        enableReinitialize
        onSubmit={(values) => handleSubmit(values)}
      >
        {({ values, setFieldValue }) => (
          <Form className="form-horizontal">
            <div className="form-group !m-0">
              <FormSection title="Custom values" isFoldable className="mt-4">
                {loadingValues && (
                  <div className="col-sm-12 p-0">
                    <InlineLoader>Loading values.yaml...</InlineLoader>
                  </div>
                )}
                {!!initialValues && (
                  <WebEditorForm
                    id="helm-app-creation-editor"
                    value={values.values}
                    onChange={(value) => setFieldValue('values', value)}
                    type="yaml"
                    data-cy="helm-app-creation-editor"
                    placeholder="Define or paste the content of your values yaml file here"
                  >
                    You can get more information about Helm values file format
                    in the{' '}
                    <a
                      href="https://helm.sh/docs/chart_template_guide/values_files/"
                      target="_blank"
                      rel="noreferrer"
                    >
                      official documentation
                    </a>
                    .
                  </WebEditorForm>
                )}
              </FormSection>
            </div>

            <FormActions
              submitLabel="Install"
              loadingText="Installing Helm chart"
              isLoading={isInstalling}
              isValid={!!namespace && !!name && !loadingValues}
              data-cy="helm-install"
            />
          </Form>
        )}
      </Formik>
    </>
  );
}
